// Copyright 2016-2018, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package engine

import (
	"bytes"
	"reflect"
	"time"

	"github.com/pulumi/pulumi/pkg/v3/resource/deploy"
	"github.com/pulumi/pulumi/sdk/v3/go/common/apitype"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag/colors"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/config"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/deepcopy"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/logging"
)

// Event represents an event generated by the engine during an operation. The underlying
// type for the `Payload` field will differ depending on the value of the `Type` field
type Event struct {
	Type    EventType
	payload interface{}
}

func NewEvent(typ EventType, payload interface{}) Event {
	ok := false
	switch typ {
	case CancelEvent:
		ok = payload == nil
	case StdoutColorEvent:
		_, ok = payload.(StdoutEventPayload)
	case DiagEvent:
		_, ok = payload.(DiagEventPayload)
	case PreludeEvent:
		_, ok = payload.(PreludeEventPayload)
	case SummaryEvent:
		_, ok = payload.(SummaryEventPayload)
	case ResourcePreEvent:
		_, ok = payload.(ResourcePreEventPayload)
	case ResourceOutputsEvent:
		_, ok = payload.(ResourceOutputsEventPayload)
	case ResourceOperationFailed:
		_, ok = payload.(ResourceOperationFailedPayload)
	case PolicyViolationEvent:
		_, ok = payload.(PolicyViolationEventPayload)
	default:
		contract.Failf("unknown event type %v", typ)
	}
	contract.Assertf(ok, "invalid payload of type %T for event type %v", payload, typ)
	return Event{
		Type:    typ,
		payload: payload,
	}
}

// EventType is the kind of event being emitted.
type EventType string

const (
	CancelEvent             EventType = "cancel"
	StdoutColorEvent        EventType = "stdoutcolor"
	DiagEvent               EventType = "diag"
	PreludeEvent            EventType = "prelude"
	SummaryEvent            EventType = "summary"
	ResourcePreEvent        EventType = "resource-pre"
	ResourceOutputsEvent    EventType = "resource-outputs"
	ResourceOperationFailed EventType = "resource-operationfailed"
	PolicyViolationEvent    EventType = "policy-violation"
)

func (e Event) Payload() interface{} {
	return deepcopy.Copy(e.payload)
}

func cancelEvent() Event {
	return Event{Type: CancelEvent}
}

// DiagEventPayload is the payload for an event with type `diag`
type DiagEventPayload struct {
	URN       resource.URN
	Prefix    string
	Message   string
	Color     colors.Colorization
	Severity  diag.Severity
	StreamID  int32
	Ephemeral bool
}

// PolicyViolationEventPayload is the payload for an event with type `policy-violation`.
type PolicyViolationEventPayload struct {
	ResourceURN       resource.URN
	Message           string
	Color             colors.Colorization
	PolicyName        string
	PolicyPackName    string
	PolicyPackVersion string
	EnforcementLevel  apitype.EnforcementLevel
	Prefix            string
}

type StdoutEventPayload struct {
	Message string
	Color   colors.Colorization
}

type PreludeEventPayload struct {
	IsPreview bool              // true if this prelude is for a plan operation
	Config    map[string]string // the keys and values for config. For encrypted config, the values may be blinded
}

type SummaryEventPayload struct {
	IsPreview       bool              // true if this summary is for a plan operation
	MaybeCorrupt    bool              // true if one or more resources may be corrupt
	Duration        time.Duration     // the duration of the entire update operation (zero values for previews)
	ResourceChanges ResourceChanges   // count of changed resources, useful for reporting
	PolicyPacks     map[string]string // {policy-pack: version} for each policy pack applied
}

type ResourceOperationFailedPayload struct {
	Metadata StepEventMetadata
	Status   resource.Status
	Steps    int
}

type ResourceOutputsEventPayload struct {
	Metadata StepEventMetadata
	Planning bool
	Debug    bool
}

type ResourcePreEventPayload struct {
	Metadata StepEventMetadata
	Planning bool
	Debug    bool
}

// StepEventMetadata contains the metadata associated with a step the engine is performing.
type StepEventMetadata struct {
	Op           deploy.StepOp                  // the operation performed by this step.
	URN          resource.URN                   // the resource URN (for before and after).
	Type         tokens.Type                    // the type affected by this step.
	Old          *StepEventStateMetadata        // the state of the resource before performing this step.
	New          *StepEventStateMetadata        // the state of the resource after performing this step.
	Res          *StepEventStateMetadata        // the latest state for the resource that is known (worst case, old).
	Keys         []resource.PropertyKey         // the keys causing replacement (only for CreateStep and ReplaceStep).
	Diffs        []resource.PropertyKey         // the keys causing diffs
	DetailedDiff map[string]plugin.PropertyDiff // the rich, structured diff
	Logical      bool                           // true if this step represents a logical operation in the program.
	Provider     string                         // the provider that performed this step.
}

// StepEventStateMetadata contains detailed metadata about a resource's state pertaining to a given step.
type StepEventStateMetadata struct {
	// State contains the raw, complete state, for this resource.
	State *resource.State
	// the resource's type.
	Type tokens.Type
	// the resource's object urn, a human-friendly, unique name for the resource.
	URN resource.URN
	// true if the resource is custom, managed by a plugin.
	Custom bool
	// true if this resource is pending deletion due to a replacement.
	Delete bool
	// the resource's unique ID, assigned by the resource provider (or blank if none/uncreated).
	ID resource.ID
	// an optional parent URN that this resource belongs to.
	Parent resource.URN
	// true to "protect" this resource (protected resources cannot be deleted).
	Protect bool
	// the resource's input properties (as specified by the program). Note: because this will cross
	// over rpc boundaries it will be slightly different than the Inputs found in resource_state.
	// Specifically, secrets will have been filtered out, and large values (like assets) will be
	// have a simple hash-based representation.  This allows clients to display this information
	// properly, without worrying about leaking sensitive data, and without having to transmit huge
	// amounts of data.
	Inputs resource.PropertyMap
	// the resource's complete output state (as returned by the resource provider).  See "Inputs"
	// for additional details about how data will be transformed before going into this map.
	Outputs resource.PropertyMap
	// the resource's provider reference
	Provider string
	// InitErrors is the set of errors encountered in the process of initializing resource (i.e.,
	// during create or update).
	InitErrors []string
}

func makeEventEmitter(events chan<- Event, update UpdateInfo) (eventEmitter, error) {
	target := update.GetTarget()
	var secrets []string
	if target != nil && target.Config.HasSecureValue() {
		for k, v := range target.Config {
			if !v.Secure() {
				continue
			}

			secureValues, err := v.SecureValues(target.Decrypter)
			if err != nil {
				return eventEmitter{}, DecryptError{
					Key: k,
					Err: err,
				}
			}
			secrets = append(secrets, secureValues...)
		}
	}

	logging.AddGlobalFilter(logging.CreateFilter(secrets, "[secret]"))

	buffer, done := make(chan Event), make(chan bool)
	go queueEvents(events, buffer, done)

	return eventEmitter{
		done: done,
		ch:   buffer,
	}, nil
}

func makeQueryEventEmitter(events chan<- Event) (eventEmitter, error) {
	buffer, done := make(chan Event), make(chan bool)

	go queueEvents(events, buffer, done)

	return eventEmitter{
		done: done,
		ch:   buffer,
	}, nil
}

type eventEmitter struct {
	done <-chan bool
	ch   chan<- Event
}

func queueEvents(events chan<- Event, buffer chan Event, done chan bool) {
	// Instead of sending to the source channel directly, buffer events to account for slow receivers.
	//
	// Buffering is done by a goroutine that concurrently receives from the senders and attempts to send events to the
	// receiver. Events that are received while waiting for the receiver to catch up are buffered in a slice.
	//
	// We do not use a buffered channel because it is empirically less likely that the goroutine reading from a
	// buffered channel will be scheduled when new data is placed in the channel.

	defer close(done)

	var queue []Event
	for {
		contract.Assert(buffer != nil)

		e, ok := <-buffer
		if !ok {
			return
		}
		queue = append(queue, e)

		// While there are events in the queue, attempt to send them to the waiting receiver. If the receiver is
		// blocked and an event is received from the event senders, stick that event in the queue.
		for len(queue) > 0 {
			select {
			case e, ok := <-buffer:
				if !ok {
					// If the event source has been closed, flush the queue.
					for _, e := range queue {
						events <- e
					}
					return
				}
				queue = append(queue, e)
			case events <- queue[0]:
				queue = queue[1:]
			}
		}
	}
}

func makeStepEventMetadata(op deploy.StepOp, step deploy.Step, debug bool) StepEventMetadata {
	contract.Assert(op == step.Op() || step.Op() == deploy.OpRefresh)

	var keys, diffs []resource.PropertyKey
	if keyer, hasKeys := step.(interface{ Keys() []resource.PropertyKey }); hasKeys {
		keys = keyer.Keys()
	}
	if differ, hasDiffs := step.(interface{ Diffs() []resource.PropertyKey }); hasDiffs {
		diffs = differ.Diffs()
	}

	var detailedDiff map[string]plugin.PropertyDiff
	if detailedDiffer, hasDetailedDiff := step.(interface {
		DetailedDiff() map[string]plugin.PropertyDiff
	}); hasDetailedDiff {
		detailedDiff = detailedDiffer.DetailedDiff()
	}

	return StepEventMetadata{
		Op:           op,
		URN:          step.URN(),
		Type:         step.Type(),
		Keys:         keys,
		Diffs:        diffs,
		DetailedDiff: detailedDiff,
		Old:          makeStepEventStateMetadata(step.Old(), debug),
		New:          makeStepEventStateMetadata(step.New(), debug),
		Res:          makeStepEventStateMetadata(step.Res(), debug),
		Logical:      step.Logical(),
		Provider:     step.Provider(),
	}
}

func makeStepEventStateMetadata(state *resource.State, debug bool) *StepEventStateMetadata {
	if state == nil {
		return nil
	}

	return &StepEventStateMetadata{
		State:      state,
		Type:       state.Type,
		URN:        state.URN,
		Custom:     state.Custom,
		Delete:     state.Delete,
		ID:         state.ID,
		Parent:     state.Parent,
		Protect:    state.Protect,
		Inputs:     filterPropertyMap(state.Inputs, debug),
		Outputs:    filterPropertyMap(state.Outputs, debug),
		Provider:   state.Provider,
		InitErrors: state.InitErrors,
	}
}

func filterPropertyMap(propertyMap resource.PropertyMap, debug bool) resource.PropertyMap {
	mappable := propertyMap.Mappable()

	var filterValue func(v interface{}) interface{}

	filterPropertyValue := func(pv resource.PropertyValue) resource.PropertyValue {
		return resource.NewPropertyValue(filterValue(pv.Mappable()))
	}

	// filter values walks unwrapped (i.e. non-PropertyValue) values and applies the filter function
	// to them recursively.  The only thing the filter actually applies to is strings.
	//
	// The return value of this function should have the same type as the input value.
	filterValue = func(v interface{}) interface{} {
		if v == nil {
			return nil
		}

		// Else, check for some known primitive types.
		switch t := v.(type) {
		case bool, int, uint, int32, uint32,
			int64, uint64, float32, float64:
			// simple types.  map over as is.
			return v
		case string:
			// have to ensure we filter out secrets.
			return logging.FilterString(t)
		case *resource.Asset:
			text := t.Text
			if text != "" {
				// we don't want to include the full text of an asset as we serialize it over as
				// events.  They represent user files and are thus are unbounded in size.  Instead,
				// we only include the text if it represents a user's serialized program code, as
				// that is something we want the receiver to see to display as part of
				// progress/diffs/etc.
				if t.IsUserProgramCode() {
					// also make sure we filter this in case there are any secrets in the code.
					text = logging.FilterString(resource.MassageIfUserProgramCodeAsset(t, debug).Text)
				} else {
					// We need to have some string here so that we preserve that this is a
					// text-asset
					text = "<stripped>"
				}
			}

			return &resource.Asset{
				Sig:  t.Sig,
				Hash: t.Hash,
				Text: text,
				Path: t.Path,
				URI:  t.URI,
			}
		case *resource.Archive:
			return &resource.Archive{
				Sig:    t.Sig,
				Hash:   t.Hash,
				Path:   t.Path,
				URI:    t.URI,
				Assets: filterValue(t.Assets).(map[string]interface{}),
			}
		case resource.Secret:
			return "[secret]"
		case resource.Computed:
			return resource.Computed{
				Element: filterPropertyValue(t.Element),
			}
		case resource.Output:
			return resource.Output{
				Element: filterPropertyValue(t.Element),
			}
		case resource.ResourceReference:
			return resource.ResourceReference{
				URN:            resource.URN(filterValue(string(t.URN)).(string)),
				ID:             resource.PropertyValue{V: filterValue(t.ID.V)},
				PackageVersion: filterValue(t.PackageVersion).(string),
			}
		}

		// Next, see if it's an array, slice, pointer or struct, and handle each accordingly.
		rv := reflect.ValueOf(v)
		switch rk := rv.Type().Kind(); rk {
		case reflect.Array, reflect.Slice:
			// If an array or slice, just create an array out of it.
			var arr []interface{}
			for i := 0; i < rv.Len(); i++ {
				arr = append(arr, filterValue(rv.Index(i).Interface()))
			}
			return arr
		case reflect.Ptr:
			if rv.IsNil() {
				return nil
			}

			v1 := filterValue(rv.Elem().Interface())
			return &v1
		case reflect.Map:
			obj := make(map[string]interface{})
			for _, key := range rv.MapKeys() {
				k := key.Interface().(string)
				v := rv.MapIndex(key).Interface()
				obj[k] = filterValue(v)
			}
			return obj
		default:
			contract.Failf("Unrecognized value type: type=%v kind=%v", rv.Type(), rk)
		}

		return nil
	}

	return resource.NewPropertyMapFromMapRepl(
		mappable, nil, /*replk*/
		func(v interface{}) (resource.PropertyValue, bool) {
			return resource.NewPropertyValue(filterValue(v)), true
		})
}

func (e *eventEmitter) Close() {
	close(e.ch)
	<-e.done
}

func (e *eventEmitter) resourceOperationFailedEvent(
	step deploy.Step, status resource.Status, steps int, debug bool) {

	contract.Requiref(e != nil, "e", "!= nil")

	e.ch <- NewEvent(ResourceOperationFailed, ResourceOperationFailedPayload{
		Metadata: makeStepEventMetadata(step.Op(), step, debug),
		Status:   status,
		Steps:    steps,
	})
}

func (e *eventEmitter) resourceOutputsEvent(op deploy.StepOp, step deploy.Step, planning bool, debug bool) {
	contract.Requiref(e != nil, "e", "!= nil")

	e.ch <- NewEvent(ResourceOutputsEvent, ResourceOutputsEventPayload{
		Metadata: makeStepEventMetadata(op, step, debug),
		Planning: planning,
		Debug:    debug,
	})
}

func (e *eventEmitter) resourcePreEvent(
	step deploy.Step, planning bool, debug bool) {

	contract.Requiref(e != nil, "e", "!= nil")

	e.ch <- NewEvent(ResourcePreEvent, ResourcePreEventPayload{
		Metadata: makeStepEventMetadata(step.Op(), step, debug),
		Planning: planning,
		Debug:    debug,
	})
}

func (e *eventEmitter) preludeEvent(isPreview bool, cfg config.Map) {
	contract.Requiref(e != nil, "e", "!= nil")

	configStringMap := make(map[string]string, len(cfg))
	for k, v := range cfg {
		keyString := k.String()
		valueString, err := v.Value(config.NewBlindingDecrypter())
		contract.AssertNoError(err)
		configStringMap[keyString] = valueString
	}

	e.ch <- NewEvent(PreludeEvent, PreludeEventPayload{
		IsPreview: isPreview,
		Config:    configStringMap,
	})
}

func (e *eventEmitter) summaryEvent(preview, maybeCorrupt bool, duration time.Duration, resourceChanges ResourceChanges,
	policyPacks map[string]string) {

	contract.Requiref(e != nil, "e", "!= nil")

	e.ch <- NewEvent(SummaryEvent, SummaryEventPayload{
		IsPreview:       preview,
		MaybeCorrupt:    maybeCorrupt,
		Duration:        duration,
		ResourceChanges: resourceChanges,
		PolicyPacks:     policyPacks,
	})
}

func (e *eventEmitter) policyViolationEvent(urn resource.URN, d plugin.AnalyzeDiagnostic) {

	contract.Requiref(e != nil, "e", "!= nil")

	// Write prefix.
	var prefix bytes.Buffer
	switch d.EnforcementLevel {
	case apitype.Mandatory:
		prefix.WriteString(colors.SpecError())
	case apitype.Advisory:
		prefix.WriteString(colors.SpecWarning())
	default:
		contract.Failf("Unrecognized diagnostic severity: %v", d)
	}

	prefix.WriteString(string(d.EnforcementLevel))
	prefix.WriteString(": ")
	prefix.WriteString(colors.Reset)

	// Write the message itself.
	var buffer bytes.Buffer
	buffer.WriteString(colors.SpecNote())

	buffer.WriteString(d.Message)

	buffer.WriteString(colors.Reset)
	buffer.WriteRune('\n')

	e.ch <- NewEvent(PolicyViolationEvent, PolicyViolationEventPayload{
		ResourceURN:       urn,
		Message:           logging.FilterString(buffer.String()),
		Color:             colors.Raw,
		PolicyName:        d.PolicyName,
		PolicyPackName:    d.PolicyPackName,
		PolicyPackVersion: d.PolicyPackVersion,
		EnforcementLevel:  d.EnforcementLevel,
		Prefix:            logging.FilterString(prefix.String()),
	})
}

func diagEvent(e *eventEmitter, d *diag.Diag, prefix, msg string, sev diag.Severity,
	ephemeral bool) {
	contract.Requiref(e != nil, "e", "!= nil")

	e.ch <- NewEvent(DiagEvent, DiagEventPayload{
		URN:       d.URN,
		Prefix:    logging.FilterString(prefix),
		Message:   logging.FilterString(msg),
		Color:     colors.Raw,
		Severity:  sev,
		StreamID:  d.StreamID,
		Ephemeral: ephemeral,
	})
}

func (e *eventEmitter) diagDebugEvent(d *diag.Diag, prefix, msg string, ephemeral bool) {
	diagEvent(e, d, prefix, msg, diag.Debug, ephemeral)
}

func (e *eventEmitter) diagInfoEvent(d *diag.Diag, prefix, msg string, ephemeral bool) {
	diagEvent(e, d, prefix, msg, diag.Info, ephemeral)
}

func (e *eventEmitter) diagInfoerrEvent(d *diag.Diag, prefix, msg string, ephemeral bool) {
	diagEvent(e, d, prefix, msg, diag.Infoerr, ephemeral)
}

func (e *eventEmitter) diagErrorEvent(d *diag.Diag, prefix, msg string, ephemeral bool) {
	diagEvent(e, d, prefix, msg, diag.Error, ephemeral)
}

func (e *eventEmitter) diagWarningEvent(d *diag.Diag, prefix, msg string, ephemeral bool) {
	diagEvent(e, d, prefix, msg, diag.Warning, ephemeral)
}
