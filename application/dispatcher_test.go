package application

import (
	"context"
	"github.com/N-Vokhmyanin/go-framework/contracts"
	"reflect"
	"testing"
)

func Test_newDispatcher(t *testing.T) {
	tests := []struct {
		name string
		want contracts.Dispatcher
	}{
		{
			name: "create new dispatcher",
			want: &dispatcher{
				listeners: make(map[string][]interface{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDispatcher(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDispatcher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_dispatcher_ListenFire(t *testing.T) {
	type event struct{}
	var catch int

	d := NewDispatcher()
	d.Listen(func(context.Context, event) {
		catch++
	})

	tests := []struct {
		name  string
		event interface{}
		want  int
	}{
		{
			name:  "first catch event",
			event: event{},
			want:  1,
		},
		{
			name:  "second catch event",
			event: event{},
			want:  2,
		},
		{
			name:  "non caught event",
			event: struct{}{},
			want:  2,
		},
		{
			name:  "third catch event",
			event: event{},
			want:  3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d.Fire(context.TODO(), tt.event)
			if tt.want != catch {
				t.Errorf("dispatcher.Fire() = %v, want %v", catch, tt.want)
			}
		})
	}
}
