package gosql

import (
	"strconv"
	"testing"
)

func Test_cleaner(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := Cleaner(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleaner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("cleaner() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readFile(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			got, err := readFile(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("readFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("readFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeSQLComments(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		args args
		want string
	}{
		// TODO: Add test cases.
		{},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if got := removeSQLComments(tt.args.content); got != tt.want {
				t.Errorf("removeSQLComments() = %v, want %v", got, tt.want)
			}
		})
	}
}
