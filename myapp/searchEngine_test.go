package myapp

import "testing"

func Test_loopGetSearchEngine(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"loopGetSearchEngine", args{
				"Boruto 81 Vostfr",
			},
			"boruto+81+vostfr",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := loopGetSearchEngine(tt.args.name); got != tt.want {
				t.Errorf("loopGetSearchEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSearchEngine(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"getSearchEngine", args{
				"narutto",
			},
			"naruto",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSearchEngine(tt.args.name); got != tt.want {
				t.Errorf("getSearchEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}