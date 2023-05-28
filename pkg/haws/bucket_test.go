package haws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"testing"
)

func TestBucket_GetExportName(t *testing.T) {
	type fields struct {
		Prefix string
	}
	type args struct {
		output string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "test-create",
			fields: fields{
				Prefix: "pref",
			},
			want: "HawsBucketPref",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bucket{
				Prefix: tt.fields.Prefix,
			}
			if got := b.GetExportName(tt.args.output); got != tt.want {
				t.Errorf("GetExportName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBucket_GetStackName(t *testing.T) {
	type fields struct {
		Prefix string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test-create",
			fields: fields{
				Prefix: "pref",
			},
			want: "pref-bucket",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bucket{
				Prefix: tt.fields.Prefix,
			}
			if got := b.GetStackName(); got != tt.want {
				t.Errorf("GetStackName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHaws_CreateBucket(t *testing.T) {
	type fields struct {
		Prefix string
		Region string
		ZoneId string
		Domain string
		Record string
		Path   string
		dryRun bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "bucket Name",
			fields: fields{
				Prefix: "pref",
				Region: "eu-east-1",
				ZoneId: "ZD1234534",
				Domain: "example.com",
				Record: "www",
				Path:   "/",
				dryRun: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Haws{
				Prefix: tt.fields.Prefix,
				Region: tt.fields.Region,
				ZoneId: tt.fields.ZoneId,
				Domain: tt.fields.Domain,
				Record: tt.fields.Record,
				Path:   tt.fields.Path,
				dryRun: tt.fields.dryRun,
			}
			bucket := h.CreateBucket("bucket")
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String("eu-central-1"),
			})
			if err != nil {
				t.Errorf("%v", err)
			}
			svc := cloudformation.New(sess)

			body, err := bucket.Build().JSON()
			if err != nil {
				t.Errorf("unable to convert template to json")
			}

			params := &cloudformation.ValidateTemplateInput{
				TemplateBody: aws.String(string(body)),
			}
			resp, err := svc.ValidateTemplate(params)

			if err != nil {
				t.Errorf("%v", err)
			}
			fmt.Println(awsutil.Prettify(resp))
		})
	}
}
