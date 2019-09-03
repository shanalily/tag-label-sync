package controller

import (
	"context"
	"fmt"
	"testing"
)

type fakeComputeResource struct {
	tags map[string]*string
}

func (c *fakeComputeResource) Update(ctx context.Context) error {
	return nil
}

func (c *fakeComputeResource) Tags() map[string]*string {
	return c.tags
}

func (c *fakeComputeResource) SetTag(name string, value *string) {
	c.tags[name] = value
}

func TestValidTagName(t *testing.T) {
	var tagNameTests = []struct {
		given    string
		expected bool
	}{
		{"kubernetes.io/arch", false},
		{"arch", true},
		{"tag?", false},
		{"To thine own self be true, and it must follow, as the night the day, thou canst not then be false to any man.", true},
		{"O, then, I see Queen Mab hath been with you. She is the fairies' midwife, and she comes In shape no bigger than an agate-stone On the fore-finger of an alderman, Drawn with a team of little atomies Athwart men's noses as they lie asleep; Her wagon-spokes made of long spinners' legs, The cover of the wings of grasshoppers, The traces of the smallest spider's web, The collars of the moonshine's watery beams, Her whip of cricket's bone, the lash of film, Her wagoner a small grey-coated gnat, Not so big as a round little worm Prick'd from the lazy finger of a maid;  Her chariot is an empty hazel-nut Made by the joiner squirrel or old grub, Time out o' mind the fairies' coachmakers.", false},
	}

	config := DefaultConfigOptions()

	for _, tt := range tagNameTests {
		t.Run(tt.given, func(t *testing.T) {
			valid := ValidTagName(tt.given, config)
			if valid != tt.expected {
				t.Errorf("given tag name %q, got valid=%t, want valid=%t", tt.given, valid, tt.expected)
			}
		})
	}
}

func TestConvertTagNameToValidLabelName(t *testing.T) {
	var tagNameConversionTests = []struct {
		given    string
		expected string
	}{
		{"env", fmt.Sprintf("%s/env", DefaultLabelPrefix)},
		{"dept", fmt.Sprintf("%s/dept", DefaultLabelPrefix)},
		{"Good_night_good_night._parting_is_such_sweet_sorrow._That_I_shall_say_good_night_till_it_be_morrow", fmt.Sprintf("%s/Good_night_good_night._parting_is_such_sweet_sorrow._That_I_sha", DefaultLabelPrefix)},
	}

	config := DefaultConfigOptions()

	for _, tt := range tagNameConversionTests {
		t.Run(tt.given, func(t *testing.T) {
			validLabelName := ConvertTagNameToValidLabelName(tt.given, config)
			if validLabelName != tt.expected {
				t.Errorf("given tag name %q, got label name %q, expected label name %q", tt.given, validLabelName, tt.expected)
			}
		})
	}
}

func TestConvertLabelNameToValidTagName(t *testing.T) {
	var labelNameConversionTests = []struct {
		given    string
		expected string
	}{
		{"favfruit", "favfruit"},
	}

	config := DefaultConfigOptions()

	for _, tt := range labelNameConversionTests {
		t.Run(tt.given, func(t *testing.T) {
			validTagName := ConvertLabelNameToValidTagName(tt.given, config)
			if validTagName != tt.expected {
				t.Errorf("given label name %q, got tag name %q, expected tag name %q", tt.given, validTagName, tt.expected)
			}
		})
	}
}

func TestConvertTagValToValidLabelVal(t *testing.T) {
	var tagValConversionTests = []struct {
		given    string
		expected string
	}{
		{"test", "test"},
	}

	for _, tt := range tagValConversionTests {
		t.Run(tt.given, func(t *testing.T) {
			validLabelVal := ConvertTagValToValidLabelVal(tt.given)
			if validLabelVal != tt.expected {
				t.Errorf("given tag name %q, got label name %q, expected label name %q", tt.given, validLabelVal, tt.expected)
			}
		})
	}
}

func TestConvertLabelValToValidTagVal(t *testing.T) {
	var labelValConversionTests = []struct {
		given    string
		expected string
	}{
		{"test", "test"},
	}

	for _, tt := range labelValConversionTests {
		t.Run(tt.given, func(t *testing.T) {
			validTagVal := ConvertLabelValToValidTagVal(tt.given)
			if validTagVal != tt.expected {
				t.Errorf("given label name %q, got tag name %q, expected tag name %q", tt.given, validTagVal, tt.expected)
			}
		})
	}
}