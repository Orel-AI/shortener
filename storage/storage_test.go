package storage

import (
	"context"
	"github.com/Orel-AI/shortener.git/config"
	"log"
	"testing"
)

func TestFindRecord(t *testing.T) {
	type args struct {
		key    string
		data   string
		ctx    context.Context
		userID int
	}
	tests := []struct {
		name    string
		args    args
		wantRes string
	}{
		{
			name: "Success test",
			args: args{
				key:    "someKey",
				data:   "someData",
				ctx:    context.Background(),
				userID: 123124123,
			},
			wantRes: "someData",
		},
	}
	storage, err := NewStorage(config.Env{FileStoragePath: "testStorage.txt"})
	if err != nil {
		log.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := uint64(tt.args.userID)
			tt.args.ctx = context.WithValue(tt.args.ctx, "UserID", j)
			storage.AddRecord(tt.args.key, tt.args.data, tt.args.ctx)
			if gotRes := storage.FindRecord(tt.args.key, tt.args.ctx); gotRes != tt.wantRes {
				t.Errorf("FindRecord() = %v, want %v", gotRes, tt.wantRes)
			}
		})
	}
}
