package matrix

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestAudioVideoInputs(t *testing.T) {
	is := is.New(t)

	matrix := New("10.66.76.171")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	inputs, err := matrix.AudioVideoInputs(ctx)
	is.NoErr(err)
	fmt.Printf("\ninputs: %+v\n", inputs)
}
