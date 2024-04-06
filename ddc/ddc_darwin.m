// #import "Darwin"
#import <Foundation/Foundation.h>
#import <IOKit/IOKitLib.h>
#import <IOKit/graphics/IOFramebufferShared.h>
#import <IOKit/graphics/IOGraphicsLib.h>
#import <IOKit/i2c/IOI2CInterface.h>
#import <CoreGraphics/CGDirectDisplay.h>
#import <CoreGraphics/CGDisplayConfiguration.h>
#import <AppKit/NSScreen.h>
#include <sys/sysctl.h>

#include "ddc_darwin.h"

#define INPUT_SWITCH 0x60
#define DDC_WAIT 10000 // depending on display this must be set to as high as 50000
#define DDC_ITERATIONS 2 // depending on display this must be set higher

extern IOAVServiceRef IOAVServiceCreateWithService(CFAllocatorRef allocator, io_service_t service);
extern IOReturn IOAVServiceReadI2C(IOAVServiceRef service, uint32_t chipAddress, uint32_t offset, void* outputBuffer, uint32_t outputBufferSize);
extern IOReturn IOAVServiceWriteI2C(IOAVServiceRef service, uint32_t chipAddress, uint32_t dataAddress, void* inputBuffer, uint32_t inputBufferSize);

bool isIntelHardware() {
    size_t size;
    cpu_type_t type;
    size = sizeof(type);
    sysctlbyname("hw.cputype", &type, &size, NULL, 0);
    return type == CPU_TYPE_X86;
}

IOAVServiceRef findDisplayM1(int index) {
    // Search for display services
    kern_return_t err;
    IOAVServiceRef displayRef;
    io_string_t path;
    // CFMutableDictionaryRef matching = IOServiceMatching("AppleCLCD2");
    CFMutableDictionaryRef matching = IOServiceMatching("DCPAVServiceProxy");
    if (!matching || CFDictionaryGetCount(matching) < 1) {
        printf("Unable to create matching dictionary for finding displays!\n");
        return 0;
    }

    io_iterator_t displays;

    kern_return_t ret = IOServiceGetMatchingServices(kIOMainPortDefault, matching, &displays);
    if (ret != KERN_SUCCESS){
        printf("Unable to find any displays!\n");
        return NULL;
    }

    int i = index;
    io_service_t display;
    while(i >= 0) {
        display = IOIteratorNext(displays);
        i--;
        if (i > 0) {
            IOObjectRelease(display);
        }
    }
    displayRef = IOAVServiceCreateWithService(kCFAllocatorDefault, display);
    IOObjectRelease(display);
    return displayRef;
}

int sendDDCM1(IOAVServiceRef display, UInt8 command, int setValue) {
    UInt8 data[6];
    memset(data, 0, sizeof(data));

    data[0] = 0x84;
    data[1] = 0x03;
    data[2] = command;
    data[3] = (setValue) >> 8;
    data[4] = setValue & 255;
    data[5] = 0x6E ^ 0x51 ^ data[0] ^ data[1] ^ data[2] ^ data[3] ^ data[4];

    for (int i = 0; i <= DDC_ITERATIONS; i++) {

        usleep(DDC_WAIT);
        IOReturn err = IOAVServiceWriteI2C(display, 0x37, 0x51, data, 6);

        if (err) {
            return err;
        }

    }

    return 0;
}

IOServiceT findDisplayIntel(int index) {
    @autoreleasepool {
        int i = index;
        CGDirectDisplayID screenNumber;
        for (NSScreen *screen in NSScreen.screens) {
            if (CGDisplayIsBuiltin(screenNumber)) continue; // Built in displays don't use DDC
            i--;
            if (i >= 0) {
                continue;
            }
            NSDictionary *description = [screen deviceDescription];
            if (![description objectForKey:@"NSDeviceIsScreen"]) {
                continue;
            }

            screenNumber = [[description objectForKey:@"NSScreenNumber"] unsignedIntValue];
        }

        io_service_t framebuffer = CGDisplayIOServicePort(screenNumber);
        return (IOServiceT) framebuffer;
    }
}

bool sendI2CDDC(IOI2CRequest *request, IOServiceT framebuffer) {
    bool result = false;
    bool sendResult = false;
    IOItemCount busCount;
    if (IOFBGetI2CInterfaceCount(framebuffer, &busCount) == KERN_SUCCESS) {
        IOOptionBits bus = 0;
        printf("Device has bus count %d\n", busCount);
        while (bus < busCount) {
            io_service_t interface;
            if (IOFBCopyI2CInterfaceForBus(framebuffer, bus++, &interface) != KERN_SUCCESS)
                continue;

            IOI2CConnectRef connect;
            if (IOI2CInterfaceOpen(interface, kNilOptions, &connect) == KERN_SUCCESS) {
                sendResult = (IOI2CSendRequest(connect, kNilOptions, request) == KERN_SUCCESS);
                IOI2CInterfaceClose(connect, kNilOptions);
            }

            if (request->replyTransactionType == kIOI2CNoTransactionType)
                usleep(20000);
            if (request->result == kIOReturnNoDevice) {
                printf("No Device for bus %d\n", bus-1);
            }
            if (request->result == kIOReturnUnsupportedMode) {
                printf("Unsupported Mode for bus %d\n", bus-1);
            }

            IOObjectRelease(interface);
            result |= sendResult && request->result == KERN_SUCCESS;
        }
    }
    return result;
}

int sendDDCIntel(IOServiceT display, unsigned char command, int setValue) {
    UInt8 data[128];
    memset(data, 0, sizeof(data));

    data[0] = 0x51;
    data[1] = 0x84;
    data[2] = 0x03;
    data[3] = command;
    data[4] = (setValue) >> 8;
    data[5] = setValue & 255;
    data[6] = 0x6E ^ data[0] ^ data[1] ^ data[2] ^ data[3] ^ data[4] ^ data[5];

    for (int i = 0; i <= DDC_ITERATIONS; i++) {

        usleep(DDC_WAIT);

        IOI2CRequest request;
        request.commFlags = 0;
        request.sendAddress = 0x6E;
        request.sendTransactionType = kIOI2CSimpleTransactionType;
        request.sendBuffer = (vm_address_t) &data[0];
        request.sendBytes = 7;
        request.replyTransactionType = kIOI2CNoTransactionType;
        request.replyBytes = 0;
        if (!sendI2CDDC(&request, display)) {
            return 1;
        }
    }

    return 0;
}