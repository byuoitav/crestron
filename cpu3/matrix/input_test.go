package matrix

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
)

const (
	testDeviceAddr       = "10.66.76.171"
	testOutputSlotStart  = 33
	testRouteOutputStart = 101
)

func TestAudioVideoInputs(t *testing.T) {
	is := is.New(t)

	matrix := New(testDeviceAddr)
	matrix.OutputSlotStart = testOutputSlotStart
	matrix.SetRouteOutputStart = testRouteOutputStart

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inputs, err := matrix.AudioVideoInputs(ctx)
	is.NoErr(err)
	is.True(len(inputs) > 0)
}

func TestSetAudioVideoInputs(t *testing.T) {
	if !testing.Short() {
		t.Skip("skipping test in full mode")
	}

	is := is.New(t)

	matrix := New(testDeviceAddr)
	matrix.OutputSlotStart = testOutputSlotStart
	matrix.SetRouteOutputStart = testRouteOutputStart

	output := fmt.Sprintf("%d", testOutputSlotStart)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	is.NoErr(matrix.SetAudioVideoInput(ctx, output, "1"))

	inputs, err := matrix.AudioVideoInputs(ctx)
	is.NoErr(err)
	is.True(inputs[output] == "1")

	is.NoErr(matrix.SetAudioVideoInput(ctx, output, "2"))

	inputs, err = matrix.AudioVideoInputs(ctx)
	is.NoErr(err)
	is.True(inputs[output] == "2")
}

func TestSetAllInputs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	is := is.New(t)

	matrix := New(testDeviceAddr)
	matrix.OutputSlotStart = testOutputSlotStart
	matrix.SetRouteOutputStart = testRouteOutputStart

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	inputs, err := matrix.AudioVideoInputs(ctx)
	is.NoErr(err)

	for output := range inputs {
		is.NoErr(matrix.SetAudioVideoInput(ctx, output, "1"))
	}

	// make sure inputs got set
	inputs, err = matrix.AudioVideoInputs(ctx)
	is.NoErr(err)

	for _, input := range inputs {
		is.True(input == "1")
	}

	// change input to 2
	for output := range inputs {
		is.NoErr(matrix.SetAudioVideoInput(ctx, output, "2"))
	}

	// make sure inputs got set
	inputs, err = matrix.AudioVideoInputs(ctx)
	is.NoErr(err)

	for _, input := range inputs {
		is.True(input == "2")
	}
}
