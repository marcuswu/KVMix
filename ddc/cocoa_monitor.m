#import <Foundation/Foundation.h>
#import <IOKit/IOKitLib.h>
#import <IOKit/graphics/IOGraphicsLib.h>
#import <CoreGraphics/CGDisplayConfiguration.h>

#include "cocoa_monitor.h"

//========================================================================
// GLFW 3.1 OS X - www.glfw.org
//------------------------------------------------------------------------
// Copyright (c) 2002-2006 Marcus Geelnard
// Copyright (c) 2006-2010 Camilla Berglund <elmindreda@elmindreda.org>
//
// This software is provided 'as-is', without any express or implied
// warranty. In no event will the authors be held liable for any damages
// arising from the use of this software.
//
// Permission is granted to anyone to use this software for any purpose,
// including commercial applications, and to alter it and redistribute it
// freely, subject to the following restrictions:
//
// 1. The origin of this software must not be misrepresented; you must not
//    claim that you wrote the original software. If you use this software
//    in a product, an acknowledgment in the product documentation would
//    be appreciated but is not required.
//
// 2. Altered source versions must be plainly marked as such, and must not
//    be misrepresented as being the original software.
//
// 3. This notice may not be removed or altered from any source
//    distribution.
//
//========================================================================

// This function has been altered from the original GLFW version based on 
// https://stackoverflow.com/a/39589275/1138868
io_service_t IOServicePortFromCGDisplayID(CGDirectDisplayID displayId) {
    io_iterator_t iter;
    io_service_t serv, servicePort = 0;

    CFMutableDictionaryRef matching = IOServiceMatching("IODisplayConnect");

    // releases matching for us
    kern_return_t err = IOServiceGetMatchingServices( kIOMasterPortDefault, matching, & iter );
    if ( err )
        return 0;

    while ( (serv = IOIteratorNext(iter)) != 0 )
    {
        CFDictionaryRef displayInfo;
        CFNumberRef vendorIDRef;
        CFNumberRef productIDRef;
        CFNumberRef serialNumberRef;

        displayInfo = IODisplayCreateInfoDictionary( serv, kIODisplayOnlyPreferredName );

        Boolean success;
        success =  CFDictionaryGetValueIfPresent( displayInfo, CFSTR(kDisplayVendorID),  (const void**) & vendorIDRef );
        success &= CFDictionaryGetValueIfPresent( displayInfo, CFSTR(kDisplayProductID), (const void**) & productIDRef );

        if ( !success )
        {
            CFRelease(displayInfo);
            continue;
        }

        SInt32 vendorID;
        CFNumberGetValue( vendorIDRef, kCFNumberSInt32Type, &vendorID );
        SInt32 productID;
        CFNumberGetValue( productIDRef, kCFNumberSInt32Type, &productID );

        // If a serial number is found, use it.
        // Otherwise serial number will be nil (= 0) which will match with the output of 'CGDisplaySerialNumber'
        SInt32 serialNumber = 0;
        if ( CFDictionaryGetValueIfPresent(displayInfo, CFSTR(kDisplaySerialNumber), (const void**) & serialNumberRef) )
        {
            CFNumberGetValue( serialNumberRef, kCFNumberSInt32Type, &serialNumber );
        }

        // If the vendor and product id along with the serial don't match
        // then we are not looking at the correct monitor.
        // NOTE: The serial number is important in cases where two monitors
        //       are the exact same.
        if( CGDisplayVendorNumber(displayId) != vendorID ||
            CGDisplayModelNumber(displayId)  != productID /*||
            CGDisplaySerialNumber(displayId) != serialNumber*/ )
        {
            CFRelease(displayInfo);
            continue;
        }

        IORegistryEntryGetParentEntry(serv, kIOServicePlane, &servicePort);
        CFRelease(displayInfo);
        break;
    }

    IOObjectRelease(iter);
    return servicePort;
}