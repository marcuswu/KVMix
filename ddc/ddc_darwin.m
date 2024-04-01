// #import "Darwin"
#import <Foundation/Foundation.h>
#import <IOKit/IOKitLib.h>

#include "ddc_darwin.h"

#define INPUT_SWITCH 0x60
#define DDC_WAIT 10000 // depending on display this must be set to as high as 50000
#define DDC_ITERATIONS 2 // depending on display this must be set higher

extern IOAVServiceRef IOAVServiceCreateWithService(CFAllocatorRef allocator, io_service_t service);
extern IOReturn IOAVServiceReadI2C(IOAVServiceRef service, uint32_t chipAddress, uint32_t offset, void* outputBuffer, uint32_t outputBufferSize);
extern IOReturn IOAVServiceWriteI2C(IOAVServiceRef service, uint32_t chipAddress, uint32_t dataAddress, void* inputBuffer, uint32_t inputBufferSize);

IOAVServiceRef findDisplay(int index) {
    // Search for display services
    kern_return_t err;
    IOAVServiceRef displayRef;
    io_string_t path;
    // CFMutableDictionaryRef matching = IOServiceMatching("AppleCLCD2");
    CFMutableDictionaryRef matching = IOServiceMatching("DCPAVServiceProxy");
    if (!matching){
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

int sendDDC(IOAVServiceRef display, UInt8 command, int setValue) {
    UInt8 data[6];
    memset(data, 0, sizeof(data));

    data[0] = 0x84;
    data[2] = command;
    data[1] = 0x03;
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