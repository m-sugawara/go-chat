package main

import (
	"testing"
	"path/filepath"
	"io/ioutil"
	"os"
)
import gomniauthtest "github.com/stretchr/gomniauth/test"

func TestAuthAvatar(t *testing.T)  {
	var authAvatar AuthAvatar
	testUser := &gomniauthtest.TestUser{}
	testUser.On("AvatarURL").Return("", ErrNoAvatarURL)
	testChatUser := &chatUser{User: testUser}
	url, err := authAvatar.AvatarURL(testChatUser)
	if err != ErrNoAvatarURL {
		t.Error("AuthAvatar.GetAvatarURL should return ErrNoAvatarURL when there is no value.")
	}
	// set Value
	testUrl := "http://url-to-avatar/"
	testUser = &gomniauthtest.TestUser{}
	testChatUser.User = testUser
	testUser.On("AvatarURL").Return(testUrl, nil)
	url, err = authAvatar.AvatarURL(testChatUser)
	if err != nil {
		t.Error("AuthAvatar.AvatarURL shouldn't return ErrNoAvatarURL when exists some value.")
	} else {
		if url != testUrl {
			t.Error("AuthAvatar.AvatarURL should return valid URL.")
		}
	}
}

func TestGravatarAvatar(t *testing.T) {
	var gravatarAvatar GravatarAvatar
	user := &chatUser{uniqueId: "abc"}
	url, err := gravatarAvatar.AvatarURL(user)
	if err != nil {
		t.Error("AuthAvatar.GetAvatarURL shouldn't return ErrNoAvatarURL when there is some value.")
	}
	if url != "//www.gravatar.com/avatar/abc" {
		t.Errorf("GravatarAvatar.AvatarURL returned %s, that is wrong.", url)
	}
}

func TestFileSystemAvatar(t *testing.T)  {
	filename := filepath.Join("avatars", "abc.jpg")
	ioutil.WriteFile(filename, []byte{}, 0777)
	defer func() { os.Remove(filename) }()

	var fileSystemAvatar FileSystemAvatar
	user := &chatUser{uniqueId: "abc"}
	url, err := fileSystemAvatar.AvatarURL(user)
	if err != nil {
		t.Error("FileSystemAvatar.AvatarURL shouldn't return ErrNoAvatarURL.")
	}
	if url != "/avatars/abc.jpg" {
		t.Errorf("FileSystemAvatar.AvatarURL returned %s, that is wrong.", url)
	}
}