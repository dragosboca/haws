package haws

import (
	"reflect"
	"testing"
)

func TestHaws_Deploy(t *testing.T) {
	type fields struct {
		Prefix      string
		Region      string
		ZoneId      string
		Domain      string
		Record      string
		Path        string
		dryRun      bool
		bucket      Part
		certificate Part
		cloudfront  Part
		user        Part
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{{}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Haws{
				Prefix:      tt.fields.Prefix,
				Region:      tt.fields.Region,
				ZoneId:      tt.fields.ZoneId,
				Domain:      tt.fields.Domain,
				Record:      tt.fields.Record,
				Path:        tt.fields.Path,
				dryRun:      tt.fields.dryRun,
				bucket:      tt.fields.bucket,
				certificate: tt.fields.certificate,
				cloudfront:  tt.fields.cloudfront,
				user:        tt.fields.user,
			}
			if err := h.Deploy(); (err != nil) != tt.wantErr {
				t.Errorf("Deploy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHaws_GetOutputs(t *testing.T) {
	type fields struct {
		Prefix      string
		Region      string
		ZoneId      string
		Domain      string
		Record      string
		Path        string
		dryRun      bool
		bucket      Part
		certificate Part
		cloudfront  Part
		user        Part
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{{}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Haws{
				Prefix:      tt.fields.Prefix,
				Region:      tt.fields.Region,
				ZoneId:      tt.fields.ZoneId,
				Domain:      tt.fields.Domain,
				Record:      tt.fields.Record,
				Path:        tt.fields.Path,
				dryRun:      tt.fields.dryRun,
				bucket:      tt.fields.bucket,
				certificate: tt.fields.certificate,
				cloudfront:  tt.fields.cloudfront,
				user:        tt.fields.user,
			}
			if err := h.GetOutputs(); (err != nil) != tt.wantErr {
				t.Errorf("GetOutputs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHaws_WithDefaults(t *testing.T) {
	type fields struct {
		Prefix      string
		Region      string
		ZoneId      string
		Domain      string
		Record      string
		Path        string
		dryRun      bool
		bucket      Part
		certificate Part
		cloudfront  Part
		user        Part
	}
	tests := []struct {
		name   string
		fields fields
		want   *Haws
	}{{}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Haws{
				Prefix:      tt.fields.Prefix,
				Region:      tt.fields.Region,
				ZoneId:      tt.fields.ZoneId,
				Domain:      tt.fields.Domain,
				Record:      tt.fields.Record,
				Path:        tt.fields.Path,
				dryRun:      tt.fields.dryRun,
				bucket:      tt.fields.bucket,
				certificate: tt.fields.certificate,
				cloudfront:  tt.fields.cloudfront,
				user:        tt.fields.user,
			}
			if got := h.WithDefaults(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithDefaults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		prefix string
		region string
		record string
		zoneId string
		path   string
		dryRun bool
	}
	tests := []struct {
		name string
		args args
		want *Haws
	}{{}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.prefix, tt.args.region, tt.args.record, tt.args.zoneId, tt.args.path, tt.args.dryRun); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getZoneDomain(t *testing.T) {
	type args struct {
		zoneId string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{{}} // TODO: Add test cases.

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getZoneDomain(tt.args.zoneId)
			if (err != nil) != tt.wantErr {
				t.Errorf("getZoneDomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getZoneDomain() got = %v, want %v", got, tt.want)
			}
		})
	}
}
