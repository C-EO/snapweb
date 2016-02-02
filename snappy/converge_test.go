/*
 * Copyright (C) 2014-2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package snappy

import (
	"errors"
	"os"

	"github.com/ubuntu-core/snappy/client"
	"github.com/ubuntu-core/snappy/snap"
	. "gopkg.in/check.v1"
	"launchpad.net/webdm/webprogress"
)

type PackagePayloadSuite struct {
	h Handler
	c *fakeSnapdClient
}

var _ = Suite(&PackagePayloadSuite{})

func (s *PackagePayloadSuite) SetUpTest(c *C) {
	os.Setenv("SNAP_APP_DATA_PATH", c.MkDir())
	s.h.statusTracker = webprogress.New()
	s.c = &fakeSnapdClient{}
	s.h.setClient(s.c)
}

func (s *PackagePayloadSuite) TestPackageNotFound(c *C) {
	s.c.err = errors.New("the snap could not be retrieved")

	_, err := s.h.packagePayload("chatroom.ogra")
	c.Assert(err, NotNil)
}

func (s *PackagePayloadSuite) TestPackage(c *C) {
	s.c.snaps = []*client.Snap{newDefaultSnap()}

	pkg, err := s.h.packagePayload("chatroom.ogra")
	c.Assert(err, IsNil)
	c.Assert(pkg, DeepEquals, snapPkg{
		ID:            "chatroom.ogra",
		Description:   "WebRTC Video chat server for Snappy",
		DownloadSize:  0,
		Icon:          "/icons/chatroom.ogra_icon.png",
		InstalledSize: 18976651,
		Name:          "chatroom",
		Origin:        "ogra",
		Status:        "installed",
		Type:          "app",
		Version:       "0.1-8",
	})
}

type PayloadSuite struct {
	h Handler
}

var _ = Suite(&PayloadSuite{})

func (s *PayloadSuite) SetUpTest(c *C) {
	os.Setenv("SNAP_APP_DATA_PATH", c.MkDir())
	s.h.statusTracker = webprogress.New()
	s.h.setClient(&fakeSnapdClient{})
}

func (s *PayloadSuite) TestPayloadWithNoServices(c *C) {
	fakeSnap := newDefaultFakePart()

	q := s.h.snapQueryToPayload(fakeSnap)

	c.Check(q.Name, Equals, fakeSnap.name)
	c.Check(q.Version, Equals, fakeSnap.version)
	c.Check(q.Status, Equals, webprogress.StatusInstalled)
	c.Check(q.Type, Equals, fakeSnap.snapType)
	c.Check(q.UIPort, Equals, uint64(0))
	c.Check(q.Icon, Equals, "/icons/camlistore.sergiusens_icon.png")
	c.Check(q.Description, Equals, fakeSnap.description)
}

func (s *PayloadSuite) TestPayloadWithServicesButNoUI(c *C) {
	s.h.setClient(&fakeSnapdClientServicesNoExternalUI{})

	fakeSnap := newDefaultFakePart()
	q := s.h.snapQueryToPayload(fakeSnap)

	c.Assert(q.Name, Equals, fakeSnap.name)
	c.Assert(q.Version, Equals, fakeSnap.version)
	c.Assert(q.Status, Equals, webprogress.StatusInstalled)
	c.Assert(q.Type, Equals, fakeSnap.snapType)
	c.Assert(q.UIPort, Equals, uint64(0))
}

func (s *PayloadSuite) TestPayloadWithServicesUI(c *C) {
	s.h.setClient(&fakeSnapdClientServicesExternalUI{})

	fakeSnap := newDefaultFakePart()
	q := s.h.snapQueryToPayload(fakeSnap)

	c.Assert(q.Name, Equals, fakeSnap.name)
	c.Assert(q.Version, Equals, fakeSnap.version)
	c.Assert(q.Status, Equals, webprogress.StatusInstalled)
	c.Assert(q.Type, Equals, fakeSnap.snapType)
	c.Assert(q.UIPort, Equals, uint64(1024))
}

func (s *PayloadSuite) TestPayloadTypeGadget(c *C) {
	s.h.setClient(&fakeSnapdClientServicesExternalUI{})

	fakeSnap := newDefaultFakePart()
	fakeSnap.snapType = snap.TypeGadget

	q := s.h.snapQueryToPayload(fakeSnap)

	c.Assert(q.Name, Equals, fakeSnap.name)
	c.Assert(q.Version, Equals, fakeSnap.version)
	c.Assert(q.Status, Equals, webprogress.StatusInstalled)
	c.Assert(q.Type, Equals, fakeSnap.snapType)
	c.Assert(q.UIPort, Equals, uint64(0))
}

func (s *PayloadSuite) TestPayloadSnapInstalling(c *C) {
	fakeSnap := newDefaultFakePart()
	fakeSnapID := fakeSnap.Name() + "." + fakeSnap.Origin()
	s.h.statusTracker.Add(fakeSnapID, webprogress.OperationInstall)

	payload := s.h.snapQueryToPayload(fakeSnap)
	c.Assert(payload.Status, Equals, webprogress.StatusInstalling)
}

type AllPackagesSuite struct {
	c *fakeSnapdClient
	h Handler
}

var _ = Suite(&AllPackagesSuite{})

func (s *AllPackagesSuite) SetUpTest(c *C) {
	os.Setenv("SNAP_APP_DATA_PATH", c.MkDir())
	s.h.statusTracker = webprogress.New()
	s.c = &fakeSnapdClient{}
	s.h.setClient(s.c)
}

func (s *AllPackagesSuite) TestNoSnaps(c *C) {
	s.c.err = errors.New("snaps could not be filtered")

	snaps, err := s.h.allPackages(client.SnapFilter{})
	c.Assert(snaps, IsNil)
	c.Assert(err, NotNil)
}

func (s *AllPackagesSuite) TestHasSnaps(c *C) {
	s.c.snaps = []*client.Snap{
		newSnap("app2"),
		newSnap("app1"),
	}

	snaps, err := s.h.allPackages(client.SnapFilter{})
	c.Assert(err, IsNil)
	c.Assert(snaps, HasLen, 2)
	c.Assert(snaps[0].Name, Equals, "app1")
	c.Assert(snaps[1].Name, Equals, "app2")
}
