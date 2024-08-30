This generator is run by ../compute.go during build.  It should not need to be run manually.
This generator takes no arguments and generates the ComputeN functions, along with their input and output structs.
These structs are not made from a json schema so that they can be generic.
the output is written to ../compute_generated.go