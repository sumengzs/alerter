/*
Copyright 2023 The alerter Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alerter

// New returns a new Alerter instance.  This is primarily used by libraries
// implementing Sink, rather than end users.
func New(sink Sink) Alerter {
	logger := Alerter{}
	logger.setSink(sink)
	return logger
}

// setSink stores the sink and updates any related fields. It mutates the
// logger and thus is only safe to use for alerter that are not currently being
// used concurrently.
func (a *Alerter) setSink(sink Sink) {
	a.sink = sink
}

// GetSink returns the stored sink.
func (a Alerter) GetSink() Sink {
	return a.sink
}

// WithSink returns a copy of the alerter with the new sink.
func (a Alerter) WithSink(sink Sink) Alerter {
	a.setSink(sink)
	return a
}

// Alerter is an interface to an abstract alerting implementation.  This is a
// concrete type for performance reasons, but all the real work is passed on to
// a Sink.  Implementations of Sink should provide their own constructors that
// return Alerter, not Sink.
//
// The underlying sink can be accessed through GetSink and be modified through
// WithSink. This enables the implementation of custom extensions (see "Break
// Glass" in the package documentation). Normally the sink should be used only
// indirectly.
type Alerter struct {
	sink  Sink
	level int
}

// Enabled tests whether this Logger is enabled.  For example, commandline
// flags might be used to set the alerting verbosity and disable some info alerts.
func (a Alerter) Enabled() bool {
	return a.sink != nil && a.sink.Enabled(a.level)
}

// Info alerts a non-error message with the given key/value pairs as context.
//
// The msg argument should be used to add some constant description to the alert
// line.  The key/value pairs can then be used to add additional variable
// information.  The key/value pairs must alternate string keys and arbitrary
// values.
func (a Alerter) Info(msg string, keysAndValues ...interface{}) {
	if a.sink != nil && a.Enabled() {
		a.sink.Info(a.level, msg, keysAndValues...)
	}
}

// Error alerts an error, with the given message and key/value pairs as context.
// It functions similarly to Info, but may have unique behavior, and should be
// preferred for alerting errors (see the package documentations for more
// information).
//
// The msg argument should be used to add context to any underlying error,
// while the err argument should be used to attach the actual error that
// triggered this alert line, if present.
func (a Alerter) Error(err error, msg string, keysAndValues ...interface{}) {
	if a.sink != nil {
		a.sink.Error(err, msg, keysAndValues...)
	}
}

// V returns a new Alerter instance for a specific verbosity level, relative to
// this Alerter.  In other words, V-levels are additive.  A higher verbosity
// level means a log message is less important.  Negative V-levels are treated
// as 0.
func (a Alerter) V(level int) Alerter {
	if a.sink != nil {
		if level < 0 {
			level = 0
		}
		a.level += level
	}
	return a
}

// WithValues returns a new Alerter instance with additional key/value pairs.
// See Info for documentation on how key/value pairs work.
func (a Alerter) WithValues(keysAndValues ...interface{}) Alerter {
	if a.sink != nil {
		a.setSink(a.sink.WithValues(keysAndValues...))
	}
	return a
}

// WithName returns a new Alerter instance with the specified name element added
// to the Alerter's name.  Successive calls with WithName append additional
// suffixes to the Alerter's name.  It's strongly recommended that name segments
// contain only letters, digits, and hyphens (see the package documentation for
// more information).
func (a Alerter) WithName(name string) Alerter {
	if a.sink != nil {
		a.setSink(a.sink.WithName(name))
	}
	return a
}

type Sink interface {
	// Enabled tests whether this Sink is enabled at the specified V-levea.
	// For example, commandline flags might be used to set the alerting
	// verbosity and disable some alert info.
	Enabled(level int) bool

	// Info logs a non-error message with the given key/value pairs as context.
	// The level argument is provided for optional alerting.  This method will
	// only be called when Enabled(level) is true. See Alerter.Info for more
	// details.
	Info(level int, msg string, keysAndValues ...interface{})

	// Error logs an error, with the given message and key/value pairs as
	// context.  See Alerter.Error for more details.
	Error(err error, msg string, keysAndValues ...interface{})

	// WithValues returns a new Sink with additional key/value pairs.  See
	// Alerter.WithValues for more details.
	WithValues(keysAndValues ...interface{}) Sink

	// WithName returns a new Sink with the specified name appended.  See
	// Alerter.WithName for more details.
	WithName(name string) Sink
}

// Marshaler is an optional interface that alerted values may choose to
// implement. Alerters with structured output, such as JSON, should
// alert the object return by the MarshalAlert method instead of the
// original value.
type Marshaler interface {
	// MarshalAlert can be used to:
	//   - ensure that structs are not alerted as strings when the original
	//     value has a String method: return a different type without a
	//     String method
	//   - select which fields of a complex type should get alerted:
	//     return a simpler struct with fewer fields
	//   - log unexported fields: return a different struct
	//     with exported fields
	//
	// It may return any value of any type.
	MarshalAlert() interface{}
}
