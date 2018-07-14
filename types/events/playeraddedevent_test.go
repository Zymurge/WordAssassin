package events

import (
	"reflect"
	"testing"
	"time"
)

func TestPlayerAddedEvent_GetTimeCreated(t *testing.T) {
	tests := []struct {
		name string
		e    *PlayerAddedEvent
		want time.Time
	}{
		{
			"test1",
			&PlayerAddedEvent{
				TimeCreated: time.Now(),
			},
			time.Unix(13,0),
		},
					// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.GetTimeCreated(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PlayerAddedEvent.GetTimeCreated() = %v, want %v", got, tt.want)
			}
		})
	}
}
