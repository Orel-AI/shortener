package storage

import (
	"context"
	"testing"
)

func TestFindRecord(t *testing.T) {
	type args struct {
		key  string
		data string
		ctx  context.Context
	}
	tests := []struct {
		name    string
		args    args
		wantRes string
	}{
		{
			name: "Success test",
			args: args{
				key:  "someKey",
				data: "someData",
				ctx:  context.Background(),
			},
			wantRes: "someData",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Initialize()
			AddRecord(tt.args.key, tt.args.data, tt.args.ctx)
			if gotRes := FindRecord(tt.args.key, tt.args.ctx); gotRes != tt.wantRes {
				t.Errorf("FindRecord() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
