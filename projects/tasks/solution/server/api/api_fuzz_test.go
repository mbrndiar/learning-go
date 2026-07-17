package api_test

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"github.com/mbrndiar/learning-go/projects/tasks/solution/server/api"
	"github.com/mbrndiar/learning-go/projects/tasks/solution/task"
)

func FuzzDecodeCreate(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`{"title":"Learn Go"}`),
		[]byte(`{"title":" Learn Go "}`),
		[]byte(`{"title":null}`),
		[]byte(`{"title":"x","title":"y"}`),
		[]byte(`{"extra":true}`),
		{},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, body []byte) {
		request := httptest.NewRequest("POST", "/tasks", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")

		input, boundaryError := api.DecodeCreate(request)
		if boundaryError == nil {
			if err := task.ValidateTitle(input.Title); err != nil {
				t.Fatalf("DecodeCreate returned invalid input %#v: %v", input, err)
			}
		}
	})
}

func FuzzDecodeUpdate(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`{"title":"Revise Go"}`),
		[]byte(`{"completed":false}`),
		[]byte(`{"title":" Revise Go ","completed":true}`),
		[]byte(`{}`),
		[]byte(`{"completed":null}`),
		[]byte(`{"title":"x","title":"y"}`),
		{},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, body []byte) {
		request := httptest.NewRequest("PATCH", "/tasks/1", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")

		input, boundaryError := api.DecodeUpdate(request)
		if boundaryError == nil {
			if err := task.ValidateUpdate(input); err != nil {
				t.Fatalf("DecodeUpdate returned invalid input %#v: %v", input, err)
			}
		}
	})
}
