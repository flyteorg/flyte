package utils

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation"
)

func TestConvertToDNS1123CompatibleString(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "flytekit-java task execution",
			args: args{"orgflyteexamplesHelloWorldTask-0"},
			want: "orgflyteexamples-hello-world-task-0",
		},
		{
			name: "good pod name",
			args: args{"t7vyqhzju1-fib-5-0"},
			want: "t7vyqhzju1-fib-5-0",
		},
		{
			name: "good pod name with dots",
			args: args{"t7v.yqh.zju1-fib-5-0"},
			want: "t7v.yqh.zju1-fib-5-0",
		},
		{
			name: "leading hyphen",
			args: args{"-t7vyqhzju1-fib-5-0"},
			want: "t7vyqhzju1-fib-5-0",
		},
		{
			name: "leading dot",
			args: args{".t7vyqhzju1-fib-5-0"},
			want: "t7vyqhzju1-fib-5-0",
		},
		{
			name: "trailing hyphen",
			args: args{"t7vyqhzju1-fib-5-0-"},
			want: "t7vyqhzju1-fib-5-0",
		},
		{
			name: "trailing dot",
			args: args{"t7vyqhzju1-fib-5-0."},
			want: "t7vyqhzju1-fib-5-0",
		},
		{
			name: "long name",
			args: args{"0123456789012345678901234567890123456789012345678901234567890123456789"},
			want: "0123456789012345678901234567890123456789012345678901234567890123456789",
		},
		{
			name: "longer than max len (253)",
			args: args{"0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789"},
			want: "0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901-fbbrvh4i",
		},
		{
			name: "very invalid name",
			args: args{"---..t7vyqhzjJcI==u1-HelloWorldTask[].-.-."},
			want: "t7vyqhzj-jc-iu1-hello-world-task",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertToDNS1123SubdomainCompatibleString(tt.args.name)
			if errs := validation.IsDNS1123Subdomain(got); len(errs) > 0 {
				t.Errorf("ConvertToDNS1123SubdomainCompatibleString() = %v, which is not DNS-1123 subdomain compatible", got)
			}
			if got != tt.want {
				t.Errorf("ConvertToDNS1123SubdomainCompatibleString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertCamelCaseToKebabCase(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "flytekit-java task execution",
			args: args{"orgflyteexamplesHelloWorldTask"},
			want: "orgflyteexamples-hello-world-task",
		},
		{
			name: "good pod name",
			args: args{"t7vyqhzju1-fib-5-0"},
			want: "t7vyqhzju1-fib-5-0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ConvertCamelCaseToKebabCase(tt.args.name); got != tt.want {
				t.Errorf("ConvertCamelCaseToKebabCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
