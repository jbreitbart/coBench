package stats

import (
	"math/bits"
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
		r := make([]DataPerRun, 10)
		for i := 0; i < 10; i++ {
			r[i].Runtime = time.Duration(i)
		}
		AddReferenceRuntime(app, r)

		for _, CAT := range getCATMasks() {
			r := make([]DataPerRun, 10)
			for i := 0; i < 10; i++ {
				r[i].Runtime = time.Duration(CAT) * time.Duration(i)
			}
			AddCATRuntime(app, CAT, r)
		}
	}

	r := make([]DataPerRun, 10)
	for i := 0; i < 10; i++ {
		r[i].Runtime = time.Duration(i * i)
	}
	AddCoSchedRuntime(apps[0], apps[1], r)

	SetCommandline(false, 2, []string{"/tmp", "/tmp2"}, [2]string{"0-2", "3-5"}, apps, false, "/sys/fs/res/", 15, "3", 0.002)
}

func verifySetup(t *testing.T, apps []string) {
	for _, app := range apps {
		r := make([]DataPerRun, 10)
		for i := 0; i < 10; i++ {
			r[i].Runtime = time.Duration(i)
		}
		runtime := newRuntimeT(NoCATMask, r)

		ref := GetReferenceRuntime(app)
		if !reflect.DeepEqual(*ref, runtime) {
			t.Errorf("Comparision failure with GetReferenceRuntime for app %v.", app)
		}

		catRs := GetCATRuntimes(app)

		for _, CAT := range getCATMasks() {
			r := make([]DataPerRun, 10)
			for i := 0; i < 10; i++ {
				r[i].Runtime = time.Duration(CAT) * time.Duration(i)
			}
			runtime := newRuntimeT(CAT, r)

			c := (*catRs)[bits.OnesCount64(CAT)]
			if !reflect.DeepEqual(c, runtime) {
				t.Errorf("Comparision failure with GetIndvCATRuntimes for app %v.", app)
			}
		}
	}

	r := make([]DataPerRun, 10)
	for i := 0; i < 10; i++ {
		r[i].Runtime = time.Duration(i * i)
	}
	runtime := newRuntimeT(NoCATMask, r)

	co := GetCoSchedRuntimes(apps[0], apps[1])
	if !reflect.DeepEqual(*co, runtime) {
		t.Errorf("Comparision failure with GetCoSchedRuntimes for apps %v.", apps)
	}

	if !reflect.DeepEqual(apps, GetAllApplications()) {
		t.Errorf("Comparision failure with GetAllApplications for apps %v - %v.", apps, GetAllApplications())
	}
}

func getCATMasks() []uint64 {
	return []uint64{1, 3, 7}
}
