// Code generated by internal/generate/tags/main.go; DO NOT EDIT.

package elb

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// ListTags lists elb service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(conn *elb.ELB, identifier string) (tftags.KeyValueTags, error) {
	input := &elb.DescribeTagsInput{
		LoadBalancerNames: aws.StringSlice([]string{identifier}),
	}

	output, err := conn.DescribeTags(input)

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.TagDescriptions[0].Tags), nil
}


// []*SERVICE.Tag handling

// TagKeys returns elb service tag keys.
func TagKeys(tags tftags.KeyValueTags) []*elb.TagKeyOnly {
	result := make([]*elb.TagKeyOnly, 0, len(tags))

	for k := range tags.Map() {
		tagKey := &elb.TagKeyOnly{
			Key: aws.String(k),
		}

		result = append(result, tagKey)
	}

	return result
}

// Tags returns elb service tags.
func Tags(tags tftags.KeyValueTags) []*elb.Tag {
	result := make([]*elb.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &elb.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from elb service tags.
func KeyValueTags(tags []*elb.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}


// UpdateTags updates elb service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(conn *elb.ELB, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &elb.RemoveTagsInput{
			LoadBalancerNames: aws.StringSlice([]string{identifier}),
			Tags:              TagKeys(removedTags.IgnoreAWS()),
		}

		_, err := conn.RemoveTags(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &elb.AddTagsInput{
			LoadBalancerNames: aws.StringSlice([]string{identifier}),
			Tags:              Tags(updatedTags.IgnoreAWS()),
		}

		_, err := conn.AddTags(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
