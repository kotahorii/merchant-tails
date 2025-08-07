#include <stdint.h>
#include <stdlib.h>
#include "_cgo_export.h"

// C wrapper functions that call Go exported functions
GDExtensionClassInstancePtr c_create_instance_wrapper(void* class_userdata) {
    return go_create_instance(class_userdata);
}

void c_free_instance_wrapper(void* class_userdata, GDExtensionClassInstancePtr instance) {
    go_free_instance(class_userdata, instance);
}

void c_method_call_wrapper(void* method_userdata, GDExtensionClassInstancePtr instance,
                          const GDExtensionVariantPtr* args, int64_t arg_count,
                          GDExtensionVariantPtr ret, int64_t* error) {
    go_method_call(method_userdata, instance, args, arg_count, ret, error);
}