
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
    mock "github.com/stretchr/testify/mock"
)

 
// NewAsyncProducer creates a new instance of AsyncProducer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAsyncProducer (t interface {
	mock.TestingT
	Cleanup(func())
}) *AsyncProducer {
	mock := &AsyncProducer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}


// AsyncProducer is an autogenerated mock type for the AsyncProducer type
type AsyncProducer struct {
	mock.Mock
}

type AsyncProducer_Expecter struct {
	mock *mock.Mock
}

func (_m *AsyncProducer) EXPECT() *AsyncProducer_Expecter {
	return &AsyncProducer_Expecter{mock: &_m.Mock}
}

 

// Input provides a mock function for the type AsyncProducer
func (_mock *AsyncProducer) Input() chan<- bool {  
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Input")
	}

		
	var r0 chan<- bool 
	if returnFunc, ok := ret.Get(0).(func() chan<- bool); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan<- bool)
		}
	} 
	return r0
}



// AsyncProducer_Input_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Input'
type AsyncProducer_Input_Call struct {
	*mock.Call
}



// Input is a helper method to define mock.On call
func (_e *AsyncProducer_Expecter) Input() *AsyncProducer_Input_Call {
	return &AsyncProducer_Input_Call{Call: _e.mock.On("Input", )}
}

func (_c *AsyncProducer_Input_Call) Run(run func()) *AsyncProducer_Input_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AsyncProducer_Input_Call) Return(boolCh chan<- bool) *AsyncProducer_Input_Call {
	_c.Call.Return(boolCh)
	return _c
}

func (_c *AsyncProducer_Input_Call) RunAndReturn(run func()chan<- bool) *AsyncProducer_Input_Call {
	_c.Call.Return(run)
	return _c
}
 

// Output provides a mock function for the type AsyncProducer
func (_mock *AsyncProducer) Output() <-chan bool {  
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Output")
	}

		
	var r0 <-chan bool 
	if returnFunc, ok := ret.Get(0).(func() <-chan bool); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan bool)
		}
	} 
	return r0
}



// AsyncProducer_Output_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Output'
type AsyncProducer_Output_Call struct {
	*mock.Call
}



// Output is a helper method to define mock.On call
func (_e *AsyncProducer_Expecter) Output() *AsyncProducer_Output_Call {
	return &AsyncProducer_Output_Call{Call: _e.mock.On("Output", )}
}

func (_c *AsyncProducer_Output_Call) Run(run func()) *AsyncProducer_Output_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AsyncProducer_Output_Call) Return(boolCh <-chan bool) *AsyncProducer_Output_Call {
	_c.Call.Return(boolCh)
	return _c
}

func (_c *AsyncProducer_Output_Call) RunAndReturn(run func()<-chan bool) *AsyncProducer_Output_Call {
	_c.Call.Return(run)
	return _c
}
 

// Whatever provides a mock function for the type AsyncProducer
func (_mock *AsyncProducer) Whatever() chan bool {  
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Whatever")
	}

		
	var r0 chan bool 
	if returnFunc, ok := ret.Get(0).(func() chan bool); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan bool)
		}
	} 
	return r0
}



// AsyncProducer_Whatever_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Whatever'
type AsyncProducer_Whatever_Call struct {
	*mock.Call
}



// Whatever is a helper method to define mock.On call
func (_e *AsyncProducer_Expecter) Whatever() *AsyncProducer_Whatever_Call {
	return &AsyncProducer_Whatever_Call{Call: _e.mock.On("Whatever", )}
}

func (_c *AsyncProducer_Whatever_Call) Run(run func()) *AsyncProducer_Whatever_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *AsyncProducer_Whatever_Call) Return(boolCh chan bool) *AsyncProducer_Whatever_Call {
	_c.Call.Return(boolCh)
	return _c
}

func (_c *AsyncProducer_Whatever_Call) RunAndReturn(run func()chan bool) *AsyncProducer_Whatever_Call {
	_c.Call.Return(run)
	return _c
}
  
