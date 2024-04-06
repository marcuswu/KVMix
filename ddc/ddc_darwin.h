#ifndef DDC_DARWIN_H
#define DDC_DARWIN_H

typedef const void *IOAVServiceRef;
typedef unsigned int IOServiceT;
bool isIntelHardware();
IOAVServiceRef findDisplayM1(int index);
int sendDDCM1(IOAVServiceRef display, unsigned char command, int setValue);
IOServiceT findDisplayIntel(int index);
int sendDDCIntel(IOServiceT display, unsigned char command, int setValue);

#endif