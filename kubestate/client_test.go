package kubestate

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClient(t *testing.T) {
	SkipConvey("When creating client", t, func() {
		c, err := NewClient(false, "/home/<user>/.kube/config")

		So(err, ShouldBeNil)
		So(c, ShouldNotBeNil)
		So(c.clientset, ShouldNotBeNil)

		pods, err := c.GetPods()
		So(err, ShouldBeNil)
		So(pods, ShouldNotBeNil)
		So(len(pods.Items), ShouldEqual, 4)
		So(pods.Items[0].Status.ContainerStatuses[0].RestartCount, ShouldEqual, 13)
	})
}
