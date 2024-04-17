#ifndef COCOA_MONITOR_H
#define COCOA_MONITOR_H

#import <IOKit/IOTypes.h>
#import <CoreGraphics/CGDirectDisplay.h>

io_service_t IOServicePortFromCGDisplayID(CGDirectDisplayID displayId);

#endif