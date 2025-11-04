/*
 * Copyright (c) 2018 Lyft. All rights reserved.
 */

package errorcollector

import "testing"

func TestIndexRange_CanMerge(t *testing.T) {
	type fields struct {
		start int
		end   int
	}
	type args struct {
		other indexRange
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"StartsBefore", fields{start: 0, end: 1}, args{other: indexRange{start: 1, end: 3}}, true},
		{"StartsAndEndsBefore", fields{start: 0, end: 0}, args{other: indexRange{start: 1, end: 3}}, true},
		{"StartsWith", fields{start: 1, end: 2}, args{other: indexRange{start: 1, end: 3}}, true},
		{"StartsAndEndsWith", fields{start: 1, end: 1}, args{other: indexRange{start: 1, end: 3}}, true},
		{"StartsBetween", fields{start: 2, end: 3}, args{other: indexRange{start: 1, end: 3}}, true},
		{"StartsBetweenEndsAfter", fields{start: 2, end: 4}, args{other: indexRange{start: 1, end: 3}}, true},
		{"StartsAfter", fields{start: 4, end: 5}, args{other: indexRange{start: 1, end: 3}}, true},
		{"EndsBefore", fields{start: -1, end: -1}, args{other: indexRange{start: 1, end: 3}}, false},
		{"StartsLater", fields{start: 5, end: 5}, args{other: indexRange{start: 1, end: 3}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := indexRange{
				start: tt.fields.start,
				end:   tt.fields.end,
			}
			if got := r.CanMerge(tt.args.other); got != tt.want {
				t.Errorf("indexRange.CanMerge() = %v, want %v", got, tt.want)
			}
		})
	}
}
