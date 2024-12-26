package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Container addiction injection container
type Container struct {
	mu        sync.RWMutex
	services  map[reflect.Type]interface{}
	factories map[reflect.Type]interface{}
}

// NewContainer creates a new DI container
func NewContainer() *Container {
	return &Container{
		services:  make(map[reflect.Type]interface{}),
		factories: make(map[reflect.Type]interface{}),
	}
}

// Register registers a service to a container
func (c *Container) Register(service interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	t := reflect.TypeOf(service)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if _, exists := c.services[t]; exists {
		return fmt.Errorf("service already registered for type: %v", t)
	}

	c.services[t] = service
	return nil
}

// RegisterFactory registers a factory to a container
func (c *Container) RegisterFactory(factory interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	t := reflect.TypeOf(factory)
	if t.Kind() != reflect.Func {
		return fmt.Errorf("factory must be a function")
	}

	if t.NumOut() != 1 && t.NumOut() != 2 {
		return fmt.Errorf("factory must return exactly one or two values (service, error)")
	}

	serviceType := t.Out(0)
	if _, exists := c.factories[serviceType]; exists {
		return fmt.Errorf("factory already registered for type: %v", serviceType)
	}

	c.factories[serviceType] = factory
	return nil
}

// Resolve resolves a service from container
func (c *Container) Resolve(target interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	targetType := targetValue.Elem().Type()

	// First check if it is registered as a direct service
	if service, exists := c.services[targetType]; exists {
		targetValue.Elem().Set(reflect.ValueOf(service))
		return nil
	}

	// See if it is a service that needs to be created with Factory
	if factory, exists := c.factories[targetType]; exists {
		factoryValue := reflect.ValueOf(factory)
		results := factoryValue.Call(nil)

		if len(results) == 2 && !results[1].IsNil() {
			return results[1].Interface().(error)
		}

		targetValue.Elem().Set(results[0])
		return nil
	}

	return fmt.Errorf("no service or factory registered for type: %v", targetType)
}

// ResolveAll resolves all services of the specified type
func (c *Container) ResolveAll(target interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("target must be a pointer to slice")
	}

	sliceType := targetValue.Elem().Type()
	elementType := sliceType.Elem()

	var services []reflect.Value

	// Collect registered services
	for t, s := range c.services {
		if t.AssignableTo(elementType) {
			services = append(services, reflect.ValueOf(s))
		}
	}

	// Create services from factories
	for t, f := range c.factories {
		if t.AssignableTo(elementType) {
			factoryValue := reflect.ValueOf(f)
			results := factoryValue.Call(nil)

			if len(results) == 2 && !results[1].IsNil() {
				return results[1].Interface().(error)
			}

			services = append(services, results[0])
		}
	}

	// Export results to slice
	result := reflect.MakeSlice(sliceType, len(services), len(services))
	for i, service := range services {
		result.Index(i).Set(service)
	}

	targetValue.Elem().Set(result)
	return nil
}

// Clear clears container
func (c *Container) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[reflect.Type]interface{})
	c.factories = make(map[reflect.Type]interface{})
}
