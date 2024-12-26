package di

import (
	"fmt"
	"reflect"
	"sync"
)

// Container represents a dependency injection container
type Container struct {
	mu        sync.RWMutex
	services  map[reflect.Type]service
	factories map[reflect.Type]factory
}

// service represents a singleton service instance
type service struct {
	value        interface{}
	dependencies []reflect.Type
	initialized  bool
}

// factory represents a factory for creating service instances
type factory struct {
	constructor  interface{}
	dependencies []reflect.Type
	scope        Scope
}

// Scope defines the lifecycle of a service
type Scope int

const (
	// Singleton services are created once and reused
	Singleton Scope = iota
	// Transient services are created each time they are requested
	Transient
	// Scoped services are created once per scope
	Scoped
)

// Option represents a container configuration option
type Option func(*Container)

// New creates a new dependency injection container
func New(options ...Option) *Container {
	c := &Container{
		services:  make(map[reflect.Type]service),
		factories: make(map[reflect.Type]factory),
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// Register registers a service with its implementation
func (c *Container) Register(iface, impl interface{}, scope Scope) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ifaceType := reflect.TypeOf(iface)
	if ifaceType.Kind() != reflect.Ptr {
		return fmt.Errorf("interface must be a pointer type")
	}

	implType := reflect.TypeOf(impl)
	if !implType.Implements(ifaceType.Elem()) {
		return fmt.Errorf("implementation does not implement interface")
	}

	dependencies, err := c.analyzeDependencies(implType)
	if err != nil {
		return fmt.Errorf("failed to analyze dependencies: %w", err)
	}

	switch scope {
	case Singleton:
		c.services[ifaceType] = service{
			value:        impl,
			dependencies: dependencies,
		}
	case Transient, Scoped:
		c.factories[ifaceType] = factory{
			constructor:  impl,
			dependencies: dependencies,
			scope:        scope,
		}
	}

	return nil
}

// RegisterFactory registers a factory function for creating service instances
func (c *Container) RegisterFactory(iface interface{}, factoryFn interface{}, scope Scope) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	ifaceType := reflect.TypeOf(iface)
	if ifaceType.Kind() != reflect.Ptr {
		return fmt.Errorf("interface must be a pointer type")
	}

	factoryType := reflect.TypeOf(factoryFn)
	if factoryType.Kind() != reflect.Func {
		return fmt.Errorf("factory must be a function")
	}

	if factoryType.NumOut() != 1 || !factoryType.Out(0).Implements(ifaceType.Elem()) {
		return fmt.Errorf("factory must return interface type")
	}

	dependencies, err := c.analyzeDependencies(factoryType)
	if err != nil {
		return fmt.Errorf("failed to analyze factory dependencies: %w", err)
	}

	c.factories[ifaceType] = factory{
		constructor:  factoryFn,
		dependencies: dependencies,
		scope:        scope,
	}

	return nil
}

// Resolve resolves a service instance
func (c *Container) Resolve(iface interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ifaceType := reflect.TypeOf(iface)
	if ifaceType.Kind() != reflect.Ptr {
		return fmt.Errorf("interface must be a pointer type")
	}

	// Check for singleton service
	if svc, ok := c.services[ifaceType]; ok {
		if !svc.initialized {
			if err := c.initializeService(&svc); err != nil {
				return err
			}
			c.services[ifaceType] = svc
		}
		reflect.ValueOf(iface).Elem().Set(reflect.ValueOf(svc.value))
		return nil
	}

	// Check for factory
	if f, ok := c.factories[ifaceType]; ok {
		instance, err := c.createInstance(f)
		if err != nil {
			return err
		}
		reflect.ValueOf(iface).Elem().Set(reflect.ValueOf(instance))
		return nil
	}

	return fmt.Errorf("no registration found for type %v", ifaceType)
}

// analyzeDependencies analyzes the dependencies of a type
func (c *Container) analyzeDependencies(t reflect.Type) ([]reflect.Type, error) {
	var dependencies []reflect.Type

	switch t.Kind() {
	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Tag.Get("inject") != "" {
				dependencies = append(dependencies, field.Type)
			}
		}
	case reflect.Func:
		for i := 0; i < t.NumIn(); i++ {
			dependencies = append(dependencies, t.In(i))
		}
	}

	return dependencies, nil
}

// initializeService initializes a service by injecting its dependencies
func (c *Container) initializeService(svc *service) error {
	val := reflect.ValueOf(svc.value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for _, depType := range svc.dependencies {
		field := val.FieldByName(depType.Name())
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		dep := reflect.New(depType).Interface()
		if err := c.Resolve(dep); err != nil {
			return fmt.Errorf("failed to resolve dependency %v: %w", depType, err)
		}

		field.Set(reflect.ValueOf(dep).Elem())
	}

	svc.initialized = true
	return nil
}

// createInstance creates a new instance using a factory
func (c *Container) createInstance(f factory) (interface{}, error) {
	args := make([]reflect.Value, len(f.dependencies))
	for i, depType := range f.dependencies {
		dep := reflect.New(depType).Interface()
		if err := c.Resolve(dep); err != nil {
			return nil, fmt.Errorf("failed to resolve dependency %v: %w", depType, err)
		}
		args[i] = reflect.ValueOf(dep).Elem()
	}

	result := reflect.ValueOf(f.constructor).Call(args)
	if len(result) != 1 {
		return nil, fmt.Errorf("factory must return exactly one value")
	}

	return result[0].Interface(), nil
}

// Example usage in comments:
/*
	// Define interfaces
	type Logger interface {
		Log(message string)
	}

	type Database interface {
		Connect() error
	}

	// Define implementations
	type ConsoleLogger struct{}

	func (l *ConsoleLogger) Log(message string) {
		fmt.Println(message)
	}

	type PostgresDB struct {
		Logger Logger `inject:""`
		Config *Config
	}

	func (db *PostgresDB) Connect() error {
		db.Logger.Log("Connecting to database...")
		return nil
	}

	// Create container
	container := di.New()

	// Register services
	container.Register((*Logger)(nil), &ConsoleLogger{}, di.Singleton)
	container.Register((*Database)(nil), &PostgresDB{}, di.Singleton)

	// Register factory
	container.RegisterFactory((*Config)(nil), func() *Config {
		return &Config{
			Host: "localhost",
			Port: 5432,
		}
	}, di.Transient)

	// Resolve and use services
	var db Database
	if err := container.Resolve(&db); err != nil {
		log.Fatal(err)
	}

	db.Connect()
*/
