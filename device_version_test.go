package blink1

import "testing"

// firmwareVersionHID replies using the same report layout as the firmware's
// 'v' handler, without opening a USB device.
type firmwareVersionHID struct {
	major byte
	minor byte
}

func (h *firmwareVersionHID) Close()                         {}
func (h *firmwareVersionHID) Write(data []byte) error        { return nil }
func (h *firmwareVersionHID) WriteFeature(data []byte) error { return nil }
func (h *firmwareVersionHID) Read(data []byte) (int, error) {
	return h.ReadFeature(data)
}
func (h *firmwareVersionHID) ReadFeature(data []byte) (int, error) {
	return copy(data, []byte{1, 'v', 0, h.major, h.minor, 0, 0, 0, 0}), nil
}

func TestGetVersion(t *testing.T) {
	for _, tc := range []struct {
		name         string
		major, minor byte
		want         int
	}{
		{name: "mk2 204", major: '2', minor: '4', want: 204},
		{name: "mk3 303", major: '3', minor: '3', want: 303},
		{name: "mk3 304", major: '3', minor: '4', want: 304},
	} {
		t.Run(tc.name, func(t *testing.T) {
			d := &Device{dev: &firmwareVersionHID{major: tc.major, minor: tc.minor}}
			got, err := d.GetVersion()
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("GetVersion() = %d, want %d", got, tc.want)
			}
		})
	}
}
