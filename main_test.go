package main

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeUnknownJsonArray(t *testing.T) {
	var reader io.Reader = strings.NewReader(`[ { "a": 1 } ]`)

	json, err := DecodeUnknownJson(reader)

	if err != nil {
		t.Errorf("DecodeUnknownJson failed with: %v", err)
		t.FailNow()
	}

	switch jt := json.(type) {
	case JsonArray:
		inner := map[string]interface {} {"a": float64(1)}
		expected := []interface {} {inner}
		actual := json.(JsonArray).Inner
		if !reflect.DeepEqual(json.(JsonArray).Inner, expected) {
			t.Errorf("Expected object to equal (%+v)%+v was: (%v)%+v", reflect.TypeOf(expected), expected, reflect.TypeOf(actual), actual)
		}
	default:
		t.Errorf("DecodeUnknownJson did not return JsonArray: %v", jt)
	}
}

func TestDecodeUnknownJsonObject(t *testing.T) {
	var reader io.Reader = strings.NewReader(`{ "a": 1 }`)

	json, err := DecodeUnknownJson(reader)

	if err != nil {
		t.Errorf("DecodeUnknownJson failed with: %v", err)
		t.FailNow()
	}

	switch jt := json.(type) {
	case JsonObject:
		expected := map[string]interface {} {"a": float64(1)}
		actual := json.(JsonObject).Inner
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("Expected object to equal (%+v)%+v was: (%v)%+v", reflect.TypeOf(expected), expected, reflect.TypeOf(actual), actual)
		}
	default:
		t.Errorf("DecodeUnknownJson did not return JsonObject: %v", jt)
	}
}

func TestDecodeUnknownJsonString(t *testing.T) {
	var reader io.Reader = strings.NewReader(`"a"`)

	_, err := DecodeUnknownJson(reader)

	if err == nil {
		t.Error("DecodeUnknownJson should have failed with an error")
		t.FailNow()
	}
}
