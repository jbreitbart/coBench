package stats

import (
	"reflect"
	"testing"
	"time"
)

func TestWriteReadIO(t *testing.T) {
	apps := []string{"test0", "test1"}
	testSetup(t, apps)
	verifySetup(t, apps)

	StoreToFile("/tmp/coBenchTest.result.json")
	ReadFromFile("/tmp/coBenchTest.result.json")

	verifySetup(t, apps)
}

func testSetup(t *testing.T, apps []string) {
	for _, app := range apps {
		r := make([]time.Duration, 10)
		for i := 0; i < 10; i++ {
			r[i] = time.Duration(i)
		}
		AddReferenceRuntime(app, r)

		for CAT := uint64(0); CAT < 16; CAT += 2 {
			AddCATRuntime(app, CAT, r)
		}
	}

	SetCommandline(false, 2, []string{"/tmp", "/tmp2"}, [2]string{"0-2", "3-5"}, apps, false, "/sys/fs/res/", 15, "3", 0.002)
}

func verifySetup(t *testing.T, apps []string) {
	for _, app := range apps {
		r := make([]time.Duration, 10)
		for i := 0; i < 10; i++ {
			r[i] = time.Duration(i)
		}
		runtime := ComputeRuntimeStats(r)

		ref := GetReferenceRuntime(app)
		if !reflect.DeepEqual(*ref, runtime) {
			t.Errorf("Comparision failure with GetReferenceRuntime for app %v.", app)
		}

		catRs := GetIndvCATRuntimes(app)
		for CAT := uint64(0); CAT < 16; CAT += 2 {
			c := (*catRs)[CAT]
			if !reflect.DeepEqual(c, runtime) {
				t.Errorf("Comparision failure with GetIndvCATRuntimes for app %v.", app)
			}
		}
	}

	if !reflect.DeepEqual(apps, GetAllApplications()) {
		t.Errorf("Comparision failure with GetAllApplications for apps %v - %v.", apps, GetAllApplications())
	}
}
