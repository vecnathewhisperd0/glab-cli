package workspace

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPollingWriter(t *testing.T) {
	fakeRenderData := func(output string) mockData {
		return mockData{data: output}
	}

	tests := []struct {
		name              string
		mockGeneratedData []mockData
		expectedOutput    string
	}{
		{
			name:              "polling writer should skip rendering if data to render is the same as previous output",
			mockGeneratedData: []mockData{fakeRenderData("a\n"), fakeRenderData("a\n")},
			expectedOutput:    "a\n",
		},
		{
			name:              "polling writer should re-render if data to render differs from the previous output",
			mockGeneratedData: []mockData{fakeRenderData("a\n"), fakeRenderData("a\n"), fakeRenderData("b\n")},
			expectedOutput:    "a\n" + "\x1b[1A\x1b[2K" /* clear line instruction */ + "b\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			generatorTerminator := mockData{err: errors.New("terminating err")}
			mockDataToUse := append(test.mockGeneratedData, generatorTerminator)

			dataGen := mockDataGenerator{toGenerate: mockDataToUse}
			outputCollector := &bytes.Buffer{}

			writer := newPollingWriter(outputCollector, 5*time.Millisecond)

			err := writer.runRenderLoop(dataGen.generator)
			if err == nil {
				t.Error("error is expected to terminate render loop")
			} else if err != generatorTerminator.err {
				t.Errorf("unexpected error returned: %s", err.Error())
			}

			assert.EqualValues(t, test.expectedOutput, outputCollector.String())
		})
	}
}

type mockData struct {
	data string
	err  error
}

type mockDataGenerator struct {
	toGenerate []mockData
}

func (g *mockDataGenerator) generator() (string, error) {
	if len(g.toGenerate) == 0 {
		panic("no more mocks available")
	}

	toReturn := g.toGenerate[0]
	g.toGenerate = g.toGenerate[1:]

	if toReturn.err != nil {
		return "", toReturn.err
	} else {
		return toReturn.data, nil
	}
}
