package secrets

import (
	"reflect"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"
)

func TestEncodeSecretGroup(t *testing.T) {
	input := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz01234567890._-/"
	encoded := encodeSecret(input)
	t.Log(input + " -> " + encoded)
	decoded, err := decodeSecret(encoded)
	assert.NoError(t, err)
	assert.Equal(t, input, decoded)
}

func TestMarshalSecretsToMapStrings(t *testing.T) {
	type args struct {
		secrets []*core.Secret
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{name: "empty", args: args{secrets: []*core.Secret{}}, want: map[string]string{}, wantErr: false},
		{name: "nil", args: args{secrets: nil}, want: map[string]string{}, wantErr: false},
		{name: "forbidden characters", args: args{secrets: []*core.Secret{
			{
				Group: ";':/\\",
			},
		}}, want: map[string]string{
			"flyte.secrets/s0": "m4zg54lqhiqceozhhixvyxbcbi",
		}, wantErr: false},
		{name: "Without group", args: args{secrets: []*core.Secret{
			{
				Key: "my_key",
			},
		}}, want: map[string]string{
			"flyte.secrets/s0": "nnsxsoraejwxsx2lmv3secq",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalSecretsToMapStrings(tt.args.secrets)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalSecretsToMapStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err != nil {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MarshalSecretsToMapStrings() got = %v, want %v", got, tt.want)
			}
		})

		t.Run(tt.name+"_unmarshal", func(t *testing.T) {
			got, err := UnmarshalStringMapToSecrets(tt.want)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalSecretsToMapStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			} else if err != nil {
				return
			}

			if tt.args.secrets != nil && !reflect.DeepEqual(got, tt.args.secrets) {
				t.Errorf("UnmarshalSecretsToMapStrings() got = %v, want %v", got, tt.args.secrets)
			}
		})
	}
}
