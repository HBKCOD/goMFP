package funscript

import (
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Keyframe represents a single point in the script
type Keyframe struct {
	At  float64 // Time in seconds
	Pos float64 // Position from 0.0 to 1.0
}

// Bookmark represents a script bookmark
type Bookmark struct {
	Name string  `json:"name"`
	Time float64 `json:"time"` // Time in seconds
}

// Chapter represents a video chapter
type Chapter struct {
	Name      string  `json:"name"`
	StartTime float64 `json:"startTime"` // Time in seconds
	EndTime   float64 `json:"endTime"`   // Time in seconds
}

// Script contains keyframes and metadata for an axis
type Script struct {
	Name      string
	Path      string
	Keyframes []Keyframe
	Bookmarks []Bookmark
	Chapters  []Chapter
}

// JSON format structs for .funscript files
type funscriptAction struct {
	At  float64 `json:"at"`  // In milliseconds
	Pos float64 `json:"pos"` // In percent (0-100)
}

type funscriptAxis struct {
	Id      string            `json:"id"`
	Actions []funscriptAction `json:"actions"`
}

type funscriptMetadata struct {
	Bookmarks []struct {
		Name string  `json:"name"`
		Time float64 `json:"time"` // In seconds (usually) or milliseconds
	} `json:"bookmarks"`
	Chapters []struct {
		Name      string  `json:"name"`
		StartTime float64 `json:"startTime"`
		EndTime   float64 `json:"endTime"`
	} `json:"chapters"`
}

type funscriptRoot struct {
	Actions  []funscriptAction  `json:"actions"`
	Axes     []funscriptAxis    `json:"axes"`
	Metadata *funscriptMetadata `json:"metadata"`
}

// ParseFunscript parses a .funscript JSON string or stream
func ParseFunscript(r io.Reader, name, path string) (*Script, map[string]*Script, error) {
	var root funscriptRoot
	dec := json.NewDecoder(r)
	if err := dec.Decode(&root); err != nil {
		return nil, nil, err
	}

	mainScript := &Script{Name: name, Path: path}
	multiScripts := make(map[string]*Script)

	// Read bookmarks & chapters
	var bookmarks []Bookmark
	var chapters []Chapter
	if root.Metadata != nil {
		for _, b := range root.Metadata.Bookmarks {
			bookmarks = append(bookmarks, Bookmark{Name: b.Name, Time: b.Time})
		}
		for _, c := range root.Metadata.Chapters {
			chapters = append(chapters, Chapter{Name: c.Name, StartTime: c.StartTime, EndTime: c.EndTime})
		}
	}

	// Helper to convert actions to Keyframes
	actionsToKeyframes := func(actions []funscriptAction) []Keyframe {
		if len(actions) == 0 {
			return nil
		}
		kfs := make([]Keyframe, 0, len(actions))
		for _, a := range actions {
			// Convert ms to seconds, percent (0-100) to 0.0-1.0
			pos := a.Pos / 100.0
			if pos < 0 {
				pos = 0
			} else if pos > 1 {
				pos = 1
			}
			kfs = append(kfs, Keyframe{
				At:  a.At / 1000.0,
				Pos: pos,
			})
		}
		// Ensure sorted by time
		sort.Slice(kfs, func(i, j int) bool {
			return kfs[i].At < kfs[j].At
		})
		return kfs
	}

	// Standard single axis actions
	if len(root.Actions) > 0 {
		mainScript.Keyframes = actionsToKeyframes(root.Actions)
		mainScript.Bookmarks = bookmarks
		mainScript.Chapters = chapters
		// In multi-axis files, root actions represent the default L0 (Stroke) axis
		if len(root.Axes) > 0 {
			multiScripts["L0"] = mainScript
		}
	}

	// Multi-axis script actions
	for _, ax := range root.Axes {
		axisName := strings.ToUpper(ax.Id)
		kfs := actionsToKeyframes(ax.Actions)
		if len(kfs) > 0 {
			multiScripts[axisName] = &Script{
				Name:      name + "_" + axisName,
				Path:      path,
				Keyframes: kfs,
				Bookmarks: bookmarks,
				Chapters:  chapters,
			}
		}
	}

	// If there are multi-axis scripts but no root actions,
	// check if one of them is L0 to treat it as main
	if len(mainScript.Keyframes) == 0 && len(multiScripts) > 0 {
		if l0, ok := multiScripts["L0"]; ok {
			mainScript = l0
		} else {
			// Choose arbitrary axis as main
			for _, s := range multiScripts {
				mainScript = s
				break
			}
		}
	}

	return mainScript, multiScripts, nil
}

// ParseCSV parses a simple CSV file with semicolon or comma separator: time_in_sec;pos_0_to_1
func ParseCSV(r io.Reader, name, path string) (*Script, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var kfs []Keyframe
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.FieldsFunc(line, func(c rune) bool {
			return c == ';' || c == ','
		})

		if len(parts) < 2 {
			continue
		}

		timeStr := strings.ReplaceAll(parts[0], ",", ".")
		posStr := strings.ReplaceAll(parts[1], ",", ".")

		time, err := strconv.ParseFloat(timeStr, 64)
		if err != nil || time < 0 {
			continue
		}

		pos, err := strconv.ParseFloat(posStr, 64)
		if err != nil {
			continue
		}
		if pos < 0 {
			pos = 0
		} else if pos > 1 {
			pos = 1
		}

		kfs = append(kfs, Keyframe{At: time, Pos: pos})
	}

	if len(kfs) == 0 {
		return nil, errors.New("no valid keyframes in CSV")
	}

	sort.Slice(kfs, func(i, j int) bool {
		return kfs[i].At < kfs[j].At
	})

	return &Script{
		Name:      name,
		Path:      path,
		Keyframes: kfs,
	}, nil
}

// LoadScriptFromFile loads a funscript or csv file
func LoadScriptFromFile(filePath string) (*Script, map[string]*Script, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, nil, err
	}

	name := info.Name()
	if strings.HasSuffix(strings.ToLower(name), ".funscript") {
		return ParseFunscript(file, name, filePath)
	} else if strings.HasSuffix(strings.ToLower(name), ".csv") {
		s, err := ParseCSV(file, name, filePath)
		if err != nil {
			return nil, nil, err
		}
		return s, nil, nil
	}

	return nil, nil, errors.New("unsupported script file format")
}

// Evaluate returns the interpolated script value (0.0 to 1.0) at the given time
func (s *Script) Evaluate(time float64, interpType string) float64 {
	if len(s.Keyframes) == 0 {
		return 0.5 // Default neutral position
	}

	n := len(s.Keyframes)
	if time <= s.Keyframes[0].At {
		return s.Keyframes[0].Pos
	}
	if time >= s.Keyframes[n-1].At {
		return s.Keyframes[n-1].Pos
	}

	// Binary search to find the keyframe index where keyframes[idx].At > time
	idx := sort.Search(n, func(i int) bool {
		return s.Keyframes[i].At > time
	})

	// Current interval is between idx-1 and idx
	p0Idx := idx - 1
	p1Idx := idx

	p0 := s.Keyframes[p0Idx]
	p1 := s.Keyframes[p1Idx]

	switch strings.ToLower(interpType) {
	case "step":
		return p0.Pos
	case "pchip":
		// Piecewise Cubic Hermite Interpolating Polynomial
		pm1 := s.getOrExtrapolate(p0Idx-1, p1, p0)
		pp1 := s.getOrExtrapolate(p1Idx+1, p0, p1)
		s0, s1 := pchipSlopes(pm1, p0, p1, pp1)
		return cubicHermite(p0.At, p0.Pos, p1.At, p1.Pos, s0, s1, time)
	case "makima":
		// Modified Akima interpolation
		pm1 := s.getOrExtrapolate(p0Idx-1, p1, p0)
		pm2 := s.getOrExtrapolate(p0Idx-2, pm1, p1)
		pp1 := s.getOrExtrapolate(p1Idx+1, p0, p1)
		pp2 := s.getOrExtrapolate(p1Idx+2, p1, pp1)
		s0, s1 := makimaSlopes(pm2, pm1, p0, p1, pp1, pp2)
		return cubicHermite(p0.At, p0.Pos, p1.At, p1.Pos, s0, s1, time)
	case "linear":
		fallthrough
	default:
		// Linear
		t := (time - p0.At) / (p1.At - p0.At)
		return p0.Pos + (p1.Pos-p0.Pos)*t
	}
}

// getOrExtrapolate retrieves keyframe at index, or extrapolates virtual keyframe if index out of bounds
func (s *Script) getOrExtrapolate(index int, ref0, ref1 Keyframe) Keyframe {
	if index >= 0 && index < len(s.Keyframes) {
		return s.Keyframes[index]
	}
	// Extrapolate point
	return Keyframe{
		At:  3*ref1.At - 2*ref0.At,
		Pos: ref1.Pos,
	}
}

// Cubic Hermite Spline calculation
func cubicHermite(x0, y0, x1, y1, s0, s1, x float64) float64 {
	d := x1 - x0
	if math.Abs(d) < 1e-9 {
		return y0
	}
	dx := x - x0
	t := dx / d
	r := 1 - t

	val := r*r*(y0*(1+2*t)+s0*dx) + t*t*(y1*(3-2*t)-d*s1*r)
	if val < 0 {
		return 0
	}
	if val > 1 {
		return 1
	}
	return val
}

// Calculate PCHIP slopes at central points (y1 and y2)
func pchipSlopes(p0, p1, p2, p3 Keyframe) (float64, float64) {
	hkm1 := p1.At - p0.At
	var dkm1 float64
	if hkm1 > 1e-9 {
		dkm1 = (p1.Pos - p0.Pos) / hkm1
	}

	hk1 := p2.At - p1.At
	var dk1 float64
	if hk1 > 1e-9 {
		dk1 = (p2.Pos - p1.Pos) / hk1
	}

	w11 := 2*hk1 + hkm1
	w12 := hk1 + 2*hkm1
	var s1 float64
	denominator1 := w11/dkm1 + w12/dk1
	if math.IsNaN(denominator1) || math.IsInf(denominator1, 0) || denominator1 == 0 || dk1*dkm1 <= 0 {
		s1 = 0
	} else {
		s1 = (w11 + w12) / denominator1
	}

	hkm2 := p2.At - p1.At
	var dkm2 float64
	if hkm2 > 1e-9 {
		dkm2 = (p2.Pos - p1.Pos) / hkm2
	}

	hk2 := p3.At - p2.At
	var dk2 float64
	if hk2 > 1e-9 {
		dk2 = (p3.Pos - p2.Pos) / hk2
	}

	w21 := 2*hk2 + hkm2
	w22 := hk2 + 2*hkm2
	var s2 float64
	denominator2 := w21/dkm2 + w22/dk2
	if math.IsNaN(denominator2) || math.IsInf(denominator2, 0) || denominator2 == 0 || dk2*dkm2 <= 0 {
		s2 = 0
	} else {
		s2 = (w21 + w22) / denominator2
	}

	return s1, s2
}

// Calculate Makima slopes at central points
func makimaSlopes(p0, p1, p2, p3, p4, p5 Keyframe) (float64, float64) {
	slope := func(k0, k1 Keyframe) float64 {
		diff := k1.At - k0.At
		if diff < 1e-9 {
			return 0
		}
		return (k1.Pos - k0.Pos) / diff
	}

	m0 := slope(p0, p1)
	m1 := slope(p1, p2)
	m2 := slope(p2, p3)
	m3 := slope(p3, p4)
	m4 := slope(p4, p5)

	w11 := math.Abs(m3-m2) + math.Abs(m3+m2)/2.0
	w12 := math.Abs(m1-m0) + math.Abs(m1+m0)/2.0
	var s1 float64
	if w11+w12 < 1e-9 {
		s1 = 0
	} else {
		s1 = (w11*m1 + w12*m2) / (w11 + w12)
	}

	w21 := math.Abs(m4-m3) + math.Abs(m4+m3)/2.0
	w22 := math.Abs(m2-m1) + math.Abs(m2+m1)/2.0
	var s2 float64
	if w21+w22 < 1e-9 {
		s2 = 0
	} else {
		s2 = (w21*m2 + w22*m3) / (w21 + w22)
	}

	return s1, s2
}
