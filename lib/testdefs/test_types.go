// Package testdefs provides type definitions for test case naming conventions.
//
// The convention is as follows:
// X_YY_ZZZ--* : <FILENAME> NUMERICAL ORDERING OF TEST CASES
//
// X <typeof test> : Unit, Integration, e2e
// YY <typeof logic> : Class, Method, Helper, Error, etc. 
// XXX <typeof checks> : should, describe, etc.,
package testdefs

// TestType defines the type of test.
type TestType string

const (
	Unit        TestType = "Unit"
	Integration TestType = "Integration"
	E2E         TestType = "E2E"
)

// LogicType defines the type of logic being tested.
type LogicType string

const (
	Class  LogicType = "Class"
	Method LogicType = "Method"
	Helper LogicType = "Helper"
	Error  LogicType = "Error"
)

// CheckType defines the type of check being performed.
type CheckType string

const (
	Should   CheckType = "should"
	Describe CheckType = "describe"
)
