package task

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"

	_ "embed"
)

//go:embed task_ir_schema.json
var embeddedTaskIRSchema []byte

const taskIRSchemaURL = "https://runtgine.dev/schemas/task-ir-v0.1.0.json"

var (
	taskIRSchemaOnce sync.Once
	taskIRSchema     *jsonschema.Schema
	taskIRSchemaErr  error
)

func taskIRCompiled() (*jsonschema.Schema, error) {
	taskIRSchemaOnce.Do(func() {
		doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(embeddedTaskIRSchema))
		if err != nil {
			taskIRSchemaErr = fmt.Errorf("task ir schema: %w", err)
			return
		}
		c := jsonschema.NewCompiler()
		c.DefaultDraft(jsonschema.Draft2020)
		c.AssertFormat()
		if err := c.AddResource(taskIRSchemaURL, doc); err != nil {
			taskIRSchemaErr = fmt.Errorf("task ir schema: %w", err)
			return
		}
		sch, err := c.Compile(taskIRSchemaURL)
		if err != nil {
			taskIRSchemaErr = fmt.Errorf("task ir schema: %w", err)
			return
		}
		taskIRSchema = sch
	})
	return taskIRSchema, taskIRSchemaErr
}

// EmbeddedTaskIRSchema returns the embedded Task IR JSON Schema bytes.
func EmbeddedTaskIRSchema() []byte {
	return append([]byte(nil), embeddedTaskIRSchema...)
}

// ValidateDocument validates raw Task IR JSON against the Task IR schema.
// Call this on file/CLI bytes before encoding/json strips unknown fields.
func ValidateDocument(raw []byte) error {
	sch, err := taskIRCompiled()
	if err != nil {
		return err
	}
	inst, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("invalid json: %w", err)
	}
	if err := sch.Validate(inst); err != nil {
		return fmt.Errorf("task ir schema: %s", trimSchemaErr(err))
	}
	return nil
}

func trimSchemaErr(err error) string {
	s := err.Error()
	s = strings.ReplaceAll(s, "\n", "; ")
	if len(s) > 512 {
		return s[:512] + "…"
	}
	return s
}
