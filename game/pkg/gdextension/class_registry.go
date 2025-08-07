package gdextension

/*
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

typedef void* GDExtensionClassInstancePtr;
typedef void* GDExtensionMethodBindPtr;
typedef void* GDExtensionObjectPtr;
typedef void* GDExtensionConstTypePtr;
typedef void* GDExtensionTypePtr;
typedef void* GDExtensionStringNamePtr;
typedef void* GDExtensionStringPtr;
typedef void* GDExtensionVariantPtr;
typedef int64_t GDExtensionInt;

// Class creation structure
typedef struct {
	uint32_t type;
	const char* class_name;
	const char* parent_class_name;
	void* class_userdata;

	void* create_instance;
	void* free_instance;

	void* notification;

	const char* to_string;
	void* reference;
	void* unreference;

	void* get_virtual;
	void* get_rid;

	void* class_userdata2;
} GDExtensionClassCreationInfo;

// Method info structure
typedef struct {
	const char* name;
	void* method_userdata;
	void* call;
	void* ptrcall;

	uint32_t method_flags;
	uint32_t return_type;
	uint32_t return_type_metadata;
	uint32_t argument_count;
	void* arguments_info;
	void* arguments_metadata;

	uint32_t default_argument_count;
	void* default_arguments;
} GDExtensionClassMethodInfo;
*/
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

// ClassRegistry manages registered GDExtension classes
type ClassRegistry struct {
	classes map[string]*ClassInfo
	methods map[string][]*MethodInfo
	mu      sync.RWMutex
}

// ClassInfo represents a registered class
type ClassInfo struct {
	Name        string
	ParentClass string
	CreateFunc  InstanceCreateFunc
	FreeFunc    InstanceFreeFunc
	Methods     []*MethodInfo
	cInfo       *C.GDExtensionClassCreationInfo
	userData    unsafe.Pointer
}

// MethodInfo represents a registered method
type MethodInfo struct {
	Name       string
	ClassOwner string
	CallFunc   MethodCallFunc
	cInfo      *C.GDExtensionClassMethodInfo
	userData   unsafe.Pointer
}

// Function types for callbacks
type (
	// InstanceCreateFunc creates a new instance of a class
	InstanceCreateFunc func(userData unsafe.Pointer) unsafe.Pointer

	// InstanceFreeFunc frees an instance of a class
	InstanceFreeFunc func(userData unsafe.Pointer, instance unsafe.Pointer)

	// MethodCallFunc handles method calls
	MethodCallFunc func(methodData unsafe.Pointer, instance unsafe.Pointer,
		args []unsafe.Pointer, ret unsafe.Pointer) error
)

var (
	globalRegistry  *ClassRegistry
	registryOnce    sync.Once
	createCallbacks map[unsafe.Pointer]InstanceCreateFunc
	freeCallbacks   map[unsafe.Pointer]InstanceFreeFunc
	methodCallbacks map[unsafe.Pointer]MethodCallFunc
	callbackMutex   sync.RWMutex
)

func init() {
	createCallbacks = make(map[unsafe.Pointer]InstanceCreateFunc)
	freeCallbacks = make(map[unsafe.Pointer]InstanceFreeFunc)
	methodCallbacks = make(map[unsafe.Pointer]MethodCallFunc)
}

// GetClassRegistry returns the global class registry
func GetClassRegistry() *ClassRegistry {
	registryOnce.Do(func() {
		globalRegistry = &ClassRegistry{
			classes: make(map[string]*ClassInfo),
			methods: make(map[string][]*MethodInfo),
		}
	})
	return globalRegistry
}

//export go_create_instance
func go_create_instance(classUserdata unsafe.Pointer) unsafe.Pointer {
	callbackMutex.RLock()
	createFunc, exists := createCallbacks[classUserdata]
	callbackMutex.RUnlock()

	if exists && createFunc != nil {
		return createFunc(classUserdata)
	}
	return nil
}

//export go_free_instance
func go_free_instance(classUserdata unsafe.Pointer, instance unsafe.Pointer) {
	callbackMutex.RLock()
	freeFunc, exists := freeCallbacks[classUserdata]
	callbackMutex.RUnlock()

	if exists && freeFunc != nil {
		freeFunc(classUserdata, instance)
	}
}

//export go_method_call
func go_method_call(methodUserdata unsafe.Pointer, instance unsafe.Pointer,
	args unsafe.Pointer, argCount C.int64_t,
	ret unsafe.Pointer, errorOut *C.int64_t) {
	callbackMutex.RLock()
	methodFunc, exists := methodCallbacks[methodUserdata]
	callbackMutex.RUnlock()

	if exists && methodFunc != nil {
		// Convert C args to Go slice
		goArgs := make([]unsafe.Pointer, int(argCount))
		if argCount > 0 && args != nil {
			// Convert C array to Go slice - simplified approach
			for i := 0; i < int(argCount); i++ {
				// Calculate offset for each pointer in the array
				offset := uintptr(i) * unsafe.Sizeof(uintptr(0))
				argPtr := *(*unsafe.Pointer)(unsafe.Pointer(uintptr(args) + offset))
				goArgs[i] = argPtr
			}
		}

		err := methodFunc(methodUserdata, instance, goArgs, ret)
		if err != nil && errorOut != nil {
			*errorOut = 1 // Set error flag
		}
	}
}

// RegisterClass registers a new GDExtension class
func (r *ClassRegistry) RegisterClass(name, parentClass string,
	createFunc InstanceCreateFunc,
	freeFunc InstanceFreeFunc) (*ClassInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.classes[name]; exists {
		return nil, fmt.Errorf("class %s already registered", name)
	}

	// Create C class info
	cName := C.CString(name)
	cParent := C.CString(parentClass)

	cInfo := (*C.GDExtensionClassCreationInfo)(C.calloc(1, C.sizeof_GDExtensionClassCreationInfo))
	cInfo.class_name = cName
	cInfo.parent_class_name = cParent

	classInfo := &ClassInfo{
		Name:        name,
		ParentClass: parentClass,
		CreateFunc:  createFunc,
		FreeFunc:    freeFunc,
		Methods:     make([]*MethodInfo, 0),
		cInfo:       cInfo,
		userData:    unsafe.Pointer(cInfo),
	}

	// Store callbacks
	callbackMutex.Lock()
	createCallbacks[classInfo.userData] = createFunc
	freeCallbacks[classInfo.userData] = freeFunc
	callbackMutex.Unlock()

	r.classes[name] = classInfo
	return classInfo, nil
}

// RegisterMethod registers a method for a class
func (r *ClassRegistry) RegisterMethod(className, methodName string,
	callFunc MethodCallFunc) (*MethodInfo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	classInfo, exists := r.classes[className]
	if !exists {
		return nil, fmt.Errorf("class %s not found", className)
	}

	// Create C method info
	cMethodName := C.CString(methodName)

	cInfo := (*C.GDExtensionClassMethodInfo)(C.calloc(1, C.sizeof_GDExtensionClassMethodInfo))
	cInfo.name = cMethodName

	methodInfo := &MethodInfo{
		Name:       methodName,
		ClassOwner: className,
		CallFunc:   callFunc,
		cInfo:      cInfo,
		userData:   unsafe.Pointer(cInfo),
	}

	// Store callback
	callbackMutex.Lock()
	methodCallbacks[methodInfo.userData] = callFunc
	callbackMutex.Unlock()

	classInfo.Methods = append(classInfo.Methods, methodInfo)
	r.methods[className] = append(r.methods[className], methodInfo)

	return methodInfo, nil
}

// GetClass returns class info by name
func (r *ClassRegistry) GetClass(name string) (*ClassInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	class, exists := r.classes[name]
	return class, exists
}

// GetClassMethods returns all methods for a class
func (r *ClassRegistry) GetClassMethods(className string) ([]*MethodInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	methods, exists := r.methods[className]
	return methods, exists
}

// Clear removes all registered classes and methods
func (r *ClassRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Free C memory for classes
	for _, class := range r.classes {
		if class.cInfo != nil {
			C.free(unsafe.Pointer(class.cInfo.class_name))
			C.free(unsafe.Pointer(class.cInfo.parent_class_name))
			C.free(unsafe.Pointer(class.cInfo))
		}
	}

	// Free C memory for methods
	for _, methodList := range r.methods {
		for _, method := range methodList {
			if method.cInfo != nil {
				C.free(unsafe.Pointer(method.cInfo.name))
				C.free(unsafe.Pointer(method.cInfo))
			}
		}
	}

	r.classes = make(map[string]*ClassInfo)
	r.methods = make(map[string][]*MethodInfo)

	callbackMutex.Lock()
	createCallbacks = make(map[unsafe.Pointer]InstanceCreateFunc)
	freeCallbacks = make(map[unsafe.Pointer]InstanceFreeFunc)
	methodCallbacks = make(map[unsafe.Pointer]MethodCallFunc)
	callbackMutex.Unlock()
}
