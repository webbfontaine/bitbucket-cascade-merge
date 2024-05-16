package main

import (
	"fmt"
	"github.com/hashicorp/go-version"
	"reflect"
	"testing"
)

func TestRepository_URL(t *testing.T) {
	type fields struct {
		Links Links
	}
	type args struct {
		protocols string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "FindHttps",
			fields: fields{
				Links: Links{
					Clone: []*Link{{
						Name: "https",
						Href: "https://git.com/winterfell.git",
					}},
				},
			},
			args:    args{protocols: "https"},
			want:    "https://git.com/winterfell.git",
			wantErr: false,
		}, {
			name: "FindDefault",
			fields: fields{
				Links: Links{
					Clone: []*Link{{
						Name: "file",
						Href: "/tmp/git/winterfell.git",
					}},
				},
			},
			args:    args{},
			want:    "/tmp/git/winterfell.git",
			wantErr: false,
		}, {
			name: "Missing",
			fields: fields{
				Links: Links{
					Clone: []*Link{{
						Name: "file",
						Href: "/tmp/git/winterfell.git",
					}},
				},
			},
			args:    args{protocols: "https"},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Repository{
				Links: tt.fields.Links,
			}
			got, err := r.URL(tt.args.protocols)
			if (err != nil) != tt.wantErr {
				t.Errorf("URL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("URL() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCascade_Next(t *testing.T) {
	type fields struct {
		Branches []string
		Current  int
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "InBound", fields: fields{Branches: []string{"release/2", "release/3"}, Current: 0}, want: "release/3"},
		{name: "OutOfBound", fields: fields{Branches: []string{"release/2", "release/3"}, Current: 1}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cascade{
				Branches: tt.fields.Branches,
				Current:  tt.fields.Current,
			}
			if got := c.Next(); got != tt.want {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCascade_Append(t *testing.T) {
	type fields struct {
		BranchNames []string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{name: "SortNumeric", fields: fields{BranchNames: []string{"release/3", "release/2"}}, want: []string{"release/2", "release/3"}},
		{name: "SortDevel", fields: fields{BranchNames: []string{"devel", "release/3"}}, want: []string{"release/3", "devel"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cascade{
				Branches: make([]string, 0),
				Current:  0,
			}
			for _, n := range tt.fields.BranchNames {
				c.Append(n)
			}
			if !reflect.DeepEqual(c.Branches, tt.want) {
				t.Errorf("Next() = %v, want %v", c.Branches, tt.want)
			}
		})
	}
}

func TestCascade_Slice(t *testing.T) {
	type fields struct {
		TargetBranch string
		BranchNames  []string
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{name: "Unbound", fields: fields{BranchNames: []string{"release/3", "release/2"}, TargetBranch: "devel"}, want: []string{}},
		{name: "BoundLast", fields: fields{BranchNames: []string{"devel", "release/3"}, TargetBranch: "devel"}, want: []string{"devel"}},
		{name: "BoundFirst", fields: fields{BranchNames: []string{"devel", "release/2", "release/3"}, TargetBranch: "release/2"}, want: []string{"release/2", "release/3", "devel"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cascade{
				Branches: make([]string, 0),
				Current:  0,
			}
			for _, n := range tt.fields.BranchNames {
				c.Append(n)
			}
			c.Slice(tt.fields.TargetBranch)
			if !reflect.DeepEqual(c.Branches, tt.want) {
				t.Errorf("Next() = %v, want %v", c.Branches, tt.want)
			}
		})
	}
}

func mustNewVersion(v string) *version.Version {
	ver, err := version.NewVersion(v)
	if err != nil {
		panic(fmt.Sprintf("Failed to create version: %s", err))
	}
	return ver
}

func Test_extractVersion(t *testing.T) {
	type args struct {
		b string
	}
	tests := []struct {
		name string
		args args
		want *version.Version
	}{
		{
			name: "valid version",
			args: args{b: "release/22.1.1"},
			want: mustNewVersion("22.1.1"),
		}, {
			name: "valid version with 'v' prefix",
			args: args{b: "release/version_22.1.1"},
			want: mustNewVersion("22.1.1"),
		}, {
			name: "valid version with 'v' prefix",
			args: args{b: "release/v22.1.1"},
			want: mustNewVersion("22.1.1"),
		}, {
			name: "valid devel branch",
			args: args{b: "devel"},
			want: mustNewVersion("99999999"),
		}, {
			name: "invalid int",
			args: args{b: "release/not-int"},
			want: mustNewVersion("0"),
		}, {
			name: "invalid format",
			args: args{b: "invalid format"},
			want: mustNewVersion("0"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractVersion(tt.args.b); !got.Equal(tt.want) {
				t.Errorf("extractVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}
