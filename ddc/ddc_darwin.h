#ifndef DDC_DARWIN_H
#define DDC_DARWIN_H

typedef const void *IOAVServiceRef;
IOAVServiceRef findDisplay(int index);
void sendDDC(IOAVServiceRef display, unsigned char command, int setValue);

#endif