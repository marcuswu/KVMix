# KVMix
KVM switch + volume mixer via SmartKnob

## Building ##
Run `go build`

## Use ##
Copy config.yml.template to config.yml. Edit the configuration as necessary for your preferences. Ensure that config.yml is in the same directory as KVMix's working directory. Then just run KVMix.

## Program Structure ##
KVMix does the following at startup:
* Creates a connection to a SmartKnob via the SmartKnob driver code which is a port of the code written by Scott Bezek.
* Generates a configuration for the SmartKnob's UI & haptic feedback
* Each screen has a ViewModel which handles messages from the SmartKnob
  * The top of the stack is always used for handling incoming messages
  * Selecting a menu option will update state, then generate a new SmartKnob configuration
  * If a selected option changes screens, a new ViewModel is pushed onto the VM stack
  * Going back a screen pops from the ViewModel stack
  * Changing ViewModels immediately queries for configuration updates

This is basically MVVM using SmartKnob configuration as the view state