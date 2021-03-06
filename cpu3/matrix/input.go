package matrix

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (m *Matrix) AudioVideoInputs(ctx context.Context) (map[string]string, error) {
	resp, err := m.sendCommand(ctx, []byte("dumpdmrouteinfo\r\n"))
	if err != nil {
		return nil, err
	}

	var lines []string

	scanner := bufio.NewScanner(bytes.NewReader(resp))
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}

	// i don't think this will ever happen from bytes.NewReader()?
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("unable to scan response: %w", err)
	}

	if len(lines) == 0 {
		return nil, fmt.Errorf("invalid response:\n%s", resp)
	}

	inputs := make(map[string]string)

	// 1st line is always "DM Routing Information for all Output cards"
	// so just skip it
	for i := 1; i < len(lines); i++ {
		// handle input lines (skip them)
		if strings.HasPrefix(lines[i], "Routing Information for Input Card at Slot ") {
			// forward to next routing information line
			for ; i < len(lines)-1; i++ {
				if strings.HasPrefix(lines[i+1], "Routing Information") || lines[i+1] == "" {
					break
				}
			}
		} else if strings.HasPrefix(lines[i], "Routing Information for Output Card at Slot ") {
			output := strings.TrimPrefix(lines[i], "Routing Information for Output Card at Slot ")
			video := ""
			audio := ""

			// read all of the output card information
			for ; i < len(lines)-1; i++ {
				switch {
				case strings.HasPrefix(lines[i], "Video Routed From Input Card at slot "):
					video = strings.TrimPrefix(lines[i], "Video Routed From Input Card at slot ")
				case strings.HasPrefix(lines[i], "Audio Routed From Input Card at slot "):
					audio = strings.TrimPrefix(lines[i], "Audio Routed From Input Card at slot ")
				}

				if strings.HasPrefix(lines[i+1], "Routing Information") || lines[i+1] == "" {
					// finished with this card
					if audio == video {
						inputs[output] = video
					} else {
						inputs[output] = ""
					}

					break
				}

			}
		} else if lines[i] == "" {
			break
		} else {
			return inputs, fmt.Errorf("unexpected line: %s", lines[i])
		}
	}

	return inputs, nil
}

func (m *Matrix) SetAudioVideoInput(ctx context.Context, output, input string) error {
	outputNum, err := strconv.Atoi(output)
	if err != nil {
		return fmt.Errorf("unable to convert output to number: %w", err)
	}

	// i do really hate crestron for this
	outputNum = outputNum - m.OutputSlotStart + m.SetRouteOutputStart
	cmd := []byte(fmt.Sprintf("setavuroute %s %d\r\n", input, outputNum))

	resp, err := m.sendCommand(ctx, cmd)
	if err != nil {
		return err
	}

	str := string(resp)

	switch {
	case strings.Contains(str, "Invalid Output"):
		return fmt.Errorf("invalid output")
	case strings.Contains(str, "Invalid Input"):
		return fmt.Errorf("invalid input")
	}

	// wait a little so that it takes effect
	time.Sleep(2 * m.pool.Delay)

	return nil
}
