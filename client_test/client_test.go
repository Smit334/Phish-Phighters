package client_test

// You MUST NOT change these default imports.  ANY additional imports may
// break the autograder and everyone will be sad.

import (
	// Some imports use an underscore to prevent the compiler from complaining
	// about unused imports.
	_ "encoding/hex"
	_ "errors"
	_ "strconv"
	"strings"
	_ "strings"
	"testing"

	// A "dot" import is used here so that the functions in the ginko and gomega
	// modules can be used without an identifier. For example, Describe() and
	// Expect() instead of ginko.Describe() and gomega.Expect().

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	userlib "github.com/cs161-staff/project2-userlib"

	"github.com/cs161-staff/project2-starter-code/client"
)

func TestSetupAndExecution(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Client Tests")
}

// ================================================
// Global Variables (feel free to add more!)
// ================================================
const defaultPassword = "password"
const emptyString = ""
const contentOne = "Bitcoin is Nick's favorite "
const contentTwo = "digital "
const contentThree = "cryptocurrency!"

// ================================================
// Describe(...) blocks help you organize your tests
// into functional categories. They can be nested into
// a tree-like structure.
// ================================================

var _ = Describe("Client Tests", func() {

	// A few user declarations that may be used for testing. Remember to initialize these before you
	// attempt to use them!
	var alice *client.User
	var bob *client.User
	var charles *client.User
	// var doris *client.User
	// var eve *client.User
	// var frank *client.User
	// var grace *client.User
	// var horace *client.User
	// var ira *client.User

	// These declarations may be useful for multi-session testing.
	var alicePhone *client.User
	var aliceLaptop *client.User
	var aliceDesktop *client.User

	var err error

	// A bunch of filenames that may be useful.
	aliceFile := "aliceFile.txt"
	bobFile := "bobFile.txt"
	charlesFile := "charlesFile.txt"
	// dorisFile := "dorisFile.txt"
	// eveFile := "eveFile.txt"
	// frankFile := "frankFile.txt"
	// graceFile := "graceFile.txt"
	// horaceFile := "horaceFile.txt"
	// iraFile := "iraFile.txt"

	BeforeEach(func() {
		// This runs before each test within this Describe block (including nested tests).
		// Here, we reset the state of Datastore and Keystore so that tests do not interfere with each other.
		// We also initialize
		userlib.DatastoreClear()
		userlib.KeystoreClear()
	})

	Describe("Basic Tests", func() {

		Specify("Basic Test: Testing InitUser/GetUser on a single user.", func() {
			userlib.DebugMsg("Initializing user Alice.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Getting user Alice.")
			aliceLaptop, err = client.GetUser("alice", defaultPassword)
			Expect(err).To(BeNil())
		})

		Specify("Basic Test: Testing Single User Store/Load/Append.", func() {
			userlib.DebugMsg("Initializing user Alice.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Storing file data: %s", contentOne)
			err = alice.StoreFile(aliceFile, []byte(contentOne))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Appending file data: %s", contentTwo)
			err = alice.AppendToFile(aliceFile, []byte(contentTwo))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Appending file data: %s", contentThree)
			err = alice.AppendToFile(aliceFile, []byte(contentThree))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Loading file...")
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))
		})

		Specify("Basic Test: Testing Create/Accept Invite Functionality with multiple users and multiple instances.", func() {
			userlib.DebugMsg("Initializing users Alice (aliceDesktop) and Bob.")
			aliceDesktop, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			bob, err = client.InitUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Getting second instance of Alice - aliceLaptop")
			aliceLaptop, err = client.GetUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("aliceDesktop storing file %s with content: %s", aliceFile, contentOne)
			err = aliceDesktop.StoreFile(aliceFile, []byte(contentOne))
			Expect(err).To(BeNil())

			userlib.DebugMsg("aliceLaptop creating invite for Bob.")
			invite, err := aliceLaptop.CreateInvitation(aliceFile, "bob")
			Expect(err).To(BeNil())

			userlib.DebugMsg("Bob accepting invite from Alice under filename %s.", bobFile)
			err = bob.AcceptInvitation("alice", invite, bobFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Bob appending to file %s, content: %s", bobFile, contentTwo)
			err = bob.AppendToFile(bobFile, []byte(contentTwo))
			Expect(err).To(BeNil())

			userlib.DebugMsg("aliceDesktop appending to file %s, content: %s", aliceFile, contentThree)
			err = aliceDesktop.AppendToFile(aliceFile, []byte(contentThree))
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that aliceDesktop sees expected file data.")
			data, err := aliceDesktop.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))

			userlib.DebugMsg("Checking that aliceLaptop sees expected file data.")
			data, err = aliceLaptop.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))

			userlib.DebugMsg("Checking that Bob sees expected file data.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))

			userlib.DebugMsg("Getting third instance of Alice - alicePhone.")
			alicePhone, err = client.GetUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that alicePhone sees Alice's changes.")
			data, err = alicePhone.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne + contentTwo + contentThree)))
		})

		Specify("Basic Test: Testing Revoke Functionality", func() {
			userlib.DebugMsg("Initializing users Alice, Bob, and Charlie.")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			bob, err = client.InitUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			charles, err = client.InitUser("charles", defaultPassword)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Alice storing file %s with content: %s", aliceFile, contentOne)
			alice.StoreFile(aliceFile, []byte(contentOne))

			userlib.DebugMsg("Alice creating invite for Bob for file %s, and Bob accepting invite under name %s.", aliceFile, bobFile)

			invite, err := alice.CreateInvitation(aliceFile, "bob")
			Expect(err).To(BeNil())

			err = bob.AcceptInvitation("alice", invite, bobFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Alice can still load the file.")
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Bob can load the file.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Bob creating invite for Charles for file %s, and Charlie accepting invite under name %s.", bobFile, charlesFile)
			invite, err = bob.CreateInvitation(bobFile, "charles")
			Expect(err).To(BeNil())

			err = charles.AcceptInvitation("bob", invite, charlesFile)
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Bob can load the file.")
			data, err = bob.LoadFile(bobFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Charles can load the file.")
			data, err = charles.LoadFile(charlesFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Alice revoking Bob's access from %s.", aliceFile)
			err = alice.RevokeAccess(aliceFile, "bob")
			Expect(err).To(BeNil())

			userlib.DebugMsg("Checking that Alice can still load the file.")
			data, err = alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal([]byte(contentOne)))

			userlib.DebugMsg("Checking that Bob/Charles lost access to the file.")
			_, err = bob.LoadFile(bobFile)
			Expect(err).ToNot(BeNil())

			_, err = charles.LoadFile(charlesFile)
			Expect(err).ToNot(BeNil())

			userlib.DebugMsg("Checking that the revoked users cannot append to the file.")
			err = bob.AppendToFile(bobFile, []byte(contentTwo))
			Expect(err).ToNot(BeNil())

			err = charles.AppendToFile(charlesFile, []byte(contentTwo))
			Expect(err).ToNot(BeNil())
		})

	})

	Describe("Student Tests", func() {
		Specify("Use Variables", func() {
			_, _, _ = alice, bob, charles
			_, _, _ = aliceFile, bobFile, charlesFile
			_, _, _ = alicePhone, aliceLaptop, aliceDesktop
			_ = err
		})

		Specify("Student Test: Testing Init User Already Exists", func() {
			userlib.DebugMsg("Student Test: Testing Init User Already Exists")
			alice, err = client.InitUser("alice", defaultPassword)

			_, err = client.InitUser("alice", defaultPassword)
			Expect(err).ToNot(BeNil())
		})

		Specify("Student Test: Successful Login", func() {
			userlib.DebugMsg("Student Test: Successful Login")
			alice, err = client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())

			aliceData, err := client.GetUser("alice", defaultPassword)
			Expect(err).To(BeNil())
			Expect(aliceData.Username).To(Equal("alice"))
		})

		Specify("Student Test: Incorrect Password", func() {
			userlib.DebugMsg("Student Test: Incorrect Password")
			alice, err = client.InitUser("alice", defaultPassword)

			_, err = client.GetUser("alice", defaultPassword+"GARBAGE")
			Expect(err).ToNot(BeNil())
		})

		Specify("Student Test: User DNE", func() {
			userlib.DebugMsg("Student Test: User DNE")

			_, err = client.GetUser("alice", defaultPassword)
			Expect(err).ToNot(BeNil())
		})

		Specify("Student Test: Check Load File", func() {
			userlib.DebugMsg("Student Test: Check File Head")
			alice, err = client.InitUser("alice", defaultPassword)
			input := []byte("HELLO WORLD ")
			storeErr := alice.StoreFile("hello.txt", []byte(input))
			Expect(storeErr).To(BeNil())
			output, loadErr := alice.LoadFile("hello.txt")
			Expect(loadErr).To(BeNil())
			Expect(input).To(Equal(output))
		})

		Specify("Student Test: Check Load File Long", func() {
			userlib.DebugMsg("Student Test: Check File Head 2")
			alice, err = client.InitUser("alice", defaultPassword)
			input := []byte(`HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO 
							HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO HELLO `)
			storeErr := alice.StoreFile("hello.txt", []byte(input))
			Expect(storeErr).To(BeNil())
			output, loadErr := alice.LoadFile("hello.txt")
			Expect(loadErr).To(BeNil())
			Expect(input).To(Equal(output))
		})

		Specify("Student Test: Check Append", func() {
			userlib.DebugMsg("Student Test: Check Append")
			alice, err = client.InitUser("alice", defaultPassword)
			input := []byte("HELLO!")
			input2 := []byte(" WORLD!")
			storeErr := alice.StoreFile("hello.txt", []byte(input))
			Expect(storeErr).To(BeNil())
			output, loadErr := alice.LoadFile("hello.txt")
			Expect(loadErr).To(BeNil())
			Expect(input).To(Equal(output))

			appendErr := alice.AppendToFile("hello.txt", input2)
			Expect(appendErr).To(BeNil())
			output, loadErr = alice.LoadFile("hello.txt")
			Expect(loadErr).To(BeNil())
			Expect(append(input, input2...)).To(Equal(output))
		})

		Specify("Student Test: Append to File DNE", func() {
			userlib.DebugMsg("Student Test: Append to File DNE")
			alice, err = client.InitUser("alice", defaultPassword)
			err = alice.AppendToFile("hello.txt", []byte("HELLO"))
			Expect(err).ToNot(BeNil())
		})

		Specify("Student Test: Check Share End to End", func() {
			userlib.DebugMsg("Student Test: Check Share End to End")
			alice, err = client.InitUser("alice", defaultPassword)
			bob, err = client.InitUser("bob", defaultPassword)
			charlie, err := client.InitUser("charlie", defaultPassword)
			danny, err := client.InitUser("danny", defaultPassword)
			Expect(err).To(BeNil())

			input := []byte("HELLO WORLD")
			storeErr := alice.StoreFile("hello.txt", input)
			Expect(storeErr).To(BeNil())

			bobInvite, inviteErr := alice.CreateInvitation("hello.txt", "bob")
			Expect(inviteErr).To(BeNil())
			charlieInvite, inviteErr := alice.CreateInvitation("hello.txt", "charlie")
			Expect(inviteErr).To(BeNil())

			acceptErr := bob.AcceptInvitation("alice", bobInvite, "HELLO.txt")
			Expect(acceptErr).To(BeNil())
			acceptErr = charlie.AcceptInvitation("alice", charlieInvite, "hi.txt")
			Expect(acceptErr).To(BeNil())

			sharedContents, loadErr := bob.LoadFile("HELLO.txt")
			Expect(loadErr).To(BeNil())
			Expect(input).To(Equal(sharedContents))

			sharedContents, loadErr = charlie.LoadFile("hi.txt")
			Expect(loadErr).To(BeNil())
			Expect(input).To(Equal(sharedContents))

			dannyInvite, inviteErr := charlie.CreateInvitation("hi.txt", "danny")
			Expect(inviteErr).To(BeNil())
			acceptErr = danny.AcceptInvitation("charlie", dannyInvite, "HI.txt")
			Expect(acceptErr).To(BeNil())

			sharedContents, loadErr = danny.LoadFile("HI.txt")
			Expect(loadErr).To(BeNil())
			Expect(input).To(Equal(sharedContents))

			revokeErr := alice.RevokeAccess("hello.txt", "charlie")
			Expect(revokeErr).To(BeNil())
			sharedContents, loadErr = charlie.LoadFile("hi.txt")
			Expect(loadErr).ToNot(BeNil())
			sharedContents, loadErr = danny.LoadFile("HI.txt")
			Expect(loadErr).ToNot(BeNil())

			contents, loadErr := alice.LoadFile("hello.txt")
			Expect(loadErr).To(BeNil())
			Expect(input).To(Equal(contents))
			sharedContents, loadErr = bob.LoadFile("HELLO.txt")
			Expect(loadErr).To(BeNil())
			Expect(input).To(Equal(sharedContents))
		})

		Specify("Student Test: Check Share By Child's Child", func() {
			userlib.DebugMsg("Student Test: Check Share By Child's Child")
			alice, err = client.InitUser("alice", defaultPassword)
			bob, err = client.InitUser("bob", defaultPassword)
			charlie, err := client.InitUser("charlie", defaultPassword)
			danny, err := client.InitUser("danny", defaultPassword)
			Expect(err).To(BeNil())

			input := []byte("HELLO WORLD")
			storeErr := alice.StoreFile("hello.txt", input)
			Expect(storeErr).To(BeNil())

			bobInvite, inviteErr := alice.CreateInvitation("hello.txt", "bob")
			Expect(inviteErr).To(BeNil())
			acceptErr := bob.AcceptInvitation("alice", bobInvite, "HELLO.txt")
			Expect(acceptErr).To(BeNil())

			charlieInvite, inviteErr := bob.CreateInvitation("HELLO.txt", "charlie")
			Expect(inviteErr).To(BeNil())
			acceptErr = charlie.AcceptInvitation("bob", charlieInvite, "hi.txt")
			Expect(acceptErr).To(BeNil())

			dannyInvite, inviteErr := charlie.CreateInvitation("hi.txt", "danny")
			Expect(inviteErr).To(BeNil())
			acceptErr = danny.AcceptInvitation("charlie", dannyInvite, "HI.txt")
			Expect(acceptErr).To(BeNil())

			sharedContents, loadErr := danny.LoadFile("HI.txt")
			Expect(loadErr).To(BeNil())
			Expect(input).To(Equal(sharedContents))

			revokeErr := alice.RevokeAccess("hello.txt", "bob")
			Expect(revokeErr).To(BeNil())

			sharedContents, loadErr = bob.LoadFile("HELLO.txt")
			Expect(loadErr).ToNot(BeNil())
			sharedContents, loadErr = charlie.LoadFile("hi.txt")
			Expect(loadErr).ToNot(BeNil())
			sharedContents, loadErr = danny.LoadFile("HI.txt")
			Expect(loadErr).ToNot(BeNil())

			contents, loadErr := alice.LoadFile("hello.txt")
			Expect(loadErr).To(BeNil())
			Expect(input).To(Equal(contents))
		})

		Specify("Student Test: Share File to User DNE", func() {
			userlib.DebugMsg("Student Test: Share File to User DNE")
			alice, err = client.InitUser("alice", defaultPassword)
			err = alice.StoreFile("hello.txt", []byte("hello"))
			_, err = alice.CreateInvitation("hello.txt", "not real person")
			Expect(err).ToNot(BeNil())
		})

		Specify("Student Test: Share File DNE to User", func() {
			userlib.DebugMsg("Student Test: Share File DNE to User")
			alice, err = client.InitUser("alice", defaultPassword)
			bob, err = client.InitUser("bob", defaultPassword)
			_, err = alice.CreateInvitation("TEST", "bob")
			Expect(err).ToNot(BeNil())
		})

		Specify("Student Test: Test Multiple Devices Revoked", func() {
			userlib.DebugMsg("Student Test: Test Multiple Devices Revoked")
			alice, err := client.InitUser("alice", defaultPassword)
			bobPhone, err := client.InitUser("bob", defaultPassword)
			charlie, err := client.InitUser("charlie", defaultPassword)
			bobTablet, err := client.GetUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			input := []byte("HELLO WORLD")
			storeErr := alice.StoreFile("hello.txt", input)
			Expect(storeErr).To(BeNil())

			bobInvite, err := alice.CreateInvitation("hello.txt", "bob")
			Expect(err).To(BeNil())
			charlieInvite, err := alice.CreateInvitation("hello.txt", "charlie")
			Expect(err).To(BeNil())

			err = bobPhone.AcceptInvitation("alice", bobInvite, "hi.txt")
			Expect(err).To(BeNil())
			err = charlie.AcceptInvitation("alice", charlieInvite, "HELLO.txt")
			Expect(err).To(BeNil())

			contents, err := bobTablet.LoadFile("hi.txt")
			Expect(err).To(BeNil())
			Expect(input).To(Equal(contents))

			contents, err = charlie.LoadFile("HELLO.txt")
			Expect(err).To(BeNil())
			Expect(input).To(Equal(contents))

			err = alice.RevokeAccess("hello.txt", "bob")
			Expect(err).To(BeNil())

			_, err = bobPhone.LoadFile("hi.txt")
			Expect(err).ToNot(BeNil())
			_, err = bobTablet.LoadFile("hi.txt")
			Expect(err).ToNot(BeNil())

			contents, err = charlie.LoadFile("HELLO.txt")
			Expect(err).To(BeNil())
			Expect(input).To(Equal(contents))
		})

		Specify("Student Test: Non-Permitted User Accessing File", func() {
			userlib.DebugMsg("Student Test: Non-Permitted User Accessing File")
			alice, _ := client.InitUser("alice", defaultPassword)
			bob, _ := client.InitUser("bob", defaultPassword)
			alice.StoreFile(aliceFile, []byte(contentOne))
			_, err := bob.LoadFile(aliceFile)
			Expect(err).ToNot(BeNil())
		})

		Specify("Student Test: Storing and Loading Large File", func() {
			userlib.DebugMsg("Student Test: Storing and Loading Large File")
			alice, _ := client.InitUser("alice", defaultPassword)
			largeContent := make([]byte, 10000000) // 1,000,0000 bytes of data
			for i := range largeContent {
				largeContent[i] = 'a' // Fill with 'a's
			}
			err := alice.StoreFile(aliceFile, largeContent)
			Expect(err).To(BeNil())
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			Expect(data).To(Equal(largeContent))
		})

		Specify("Student Test: Multiple Users Appending to Shared File", func() {
			userlib.DebugMsg("Student Test: Multiple Users Appending to Shared File")
			alice, _ := client.InitUser("alice", defaultPassword)
			bob, _ := client.InitUser("bob", defaultPassword)
			// Alice stores the file and shares it with Bob
			alice.StoreFile(aliceFile, []byte(contentOne))
			invite, _ := alice.CreateInvitation(aliceFile, "bob")
			bob.AcceptInvitation("alice", invite, bobFile)
			// Alice appends to the file
			alice.AppendToFile(aliceFile, []byte("Alice's addition"))
			// Bob appends to the file using the shared filename
			bob.AppendToFile(bobFile, []byte("Bob's addition"))
			// Verify that both additions are present in the file
			data, _ := alice.LoadFile(aliceFile)
			Expect(data).To(ContainSubstring("Alice's addition"))
			Expect(data).To(ContainSubstring("Bob's addition"))
		})

		Specify("Student Test: InitUser with Empty Username and Password", func() {
			userlib.DebugMsg("Student Test: InitUser with Empty Username and Password")
			_, err := client.InitUser("", "")
			Expect(err).ToNot(BeNil())
		})

		Specify("File Content Integrity Test After Multiple Appends", func() {
			userlib.DebugMsg("File Content Integrity Test After Multiple Appends")
			alice, err := client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())
			err = alice.StoreFile(aliceFile, []byte(contentOne))
			Expect(err).To(BeNil())
			for i := 0; i < 10; i++ {
				err = alice.AppendToFile(aliceFile, []byte(contentTwo))
				Expect(err).To(BeNil())
			}
			data, err := alice.LoadFile(aliceFile)
			Expect(err).To(BeNil())
			expectedContent := contentOne + strings.Repeat(contentTwo, 10)
			Expect(string(data)).To(Equal(expectedContent))
		})

		Specify("Student Test: Covering Multiple Functionalities and Edge Cases", func() {
			userlib.DebugMsg("Student Test: Covering Multiple Functionalities and Edge Cases")
			// User initialization and edge cases
			alice, err := client.InitUser("alice", defaultPassword)
			Expect(err).To(BeNil())
			_, err = client.InitUser("alice", defaultPassword) // Duplicate user
			Expect(err).ToNot(BeNil())

			bob, err := client.InitUser("bob", defaultPassword)
			Expect(err).To(BeNil())

			_, err = client.InitUser("", defaultPassword) // Invalid username
			Expect(err).ToNot(BeNil())

			// Store, load, and append operations
			initialContent := []byte("Alice's initial content")
			err = alice.StoreFile("file.txt", initialContent)
			Expect(err).To(BeNil())

			loadedContent, err := alice.LoadFile("file.txt")
			Expect(err).To(BeNil())
			Expect(loadedContent).To(Equal(initialContent))

			appendContent := []byte(" - Appended by Alice")
			err = alice.AppendToFile("file.txt", appendContent)
			Expect(err).To(BeNil())

			updatedContent, err := alice.LoadFile("file.txt")
			Expect(err).To(BeNil())
			Expect(updatedContent).To(Equal(append(loadedContent, appendContent...)))

			// Sharing and accepting invitations
			bobInvite, err := alice.CreateInvitation("file.txt", "bob")
			Expect(err).To(BeNil())
			err = bob.AcceptInvitation("alice", bobInvite, "bobFile.txt")
			Expect(err).To(BeNil())

			// Bob's append operation
			bobAppendContent := []byte(" - Appended by Bob")
			err = bob.AppendToFile("bobFile.txt", bobAppendContent)
			Expect(err).To(BeNil())

			// Verify updates after sharing
			aliceUpdatedContent, err := alice.LoadFile("file.txt")
			Expect(err).To(BeNil())
			Expect(aliceUpdatedContent).To(Equal(append(updatedContent, bobAppendContent...)))

			// Revoking access and its effects
			err = alice.RevokeAccess("file.txt", "bob")
			Expect(err).To(BeNil())

			_, err = bob.LoadFile("bobFile.txt") // Bob should no longer have access
			Expect(err).ToNot(BeNil())

			// Error handling and negative test cases
			_, err = alice.LoadFile("nonexistent.txt") // Non-existent file
			Expect(err).ToNot(BeNil())

			_, err = alice.CreateInvitation("file.txt", "nonexistentuser") // Non-existent user
			Expect(err).ToNot(BeNil())

			_, err = alice.CreateInvitation("nonexistent.txt", "bob") // Non-existent file
			Expect(err).ToNot(BeNil())
		})

		Specify("Student Test: Attempt to Share File Without Ownership", func() {
			userlib.DebugMsg("Student Test: Attempt to Share File Without Ownership")
			alice, _ := client.InitUser("alice", defaultPassword)
			bob, _ := client.InitUser("bob", defaultPassword)
			alice.CreateInvitation(aliceFile, "charlie")
			Expect(err).ToNot(BeNil())
			_, err := bob.CreateInvitation(aliceFile, "charlie")
			Expect(err).ToNot(BeNil()) // Bob should not be able to share Alice's file
		})

		Specify("Student Test: Empty File Content and Filename Boundary Cases", func() {
			userlib.DebugMsg("Student Test: Empty File Content and Filename Boundary Cases")
			alice, _ := client.InitUser("alice", defaultPassword)
			err := alice.StoreFile("", []byte("")) // Empty filename and content
			Expect(err).To(BeNil())
			longFilename := strings.Repeat("d", 1000) // Very long filename
			err = alice.StoreFile(longFilename, []byte(contentOne))
			Expect(err).To(BeNil()) // Handle according to your system's design
		})

		Specify("Student Test: Unauthorized Data Access or Tampering", func() {
			userlib.DebugMsg("Student Test: Unauthorized Data Access or Tampering")
			alice, _ := client.InitUser("alice", defaultPassword)
			bob, _ := client.InitUser("bob", defaultPassword)
			alice.StoreFile(aliceFile, []byte(contentOne))
			_, err := bob.LoadFile(aliceFile)
			Expect(err).ToNot(BeNil()) // Bob should not be able to access or tamper with Alice's file
			err = bob.AppendToFile(aliceFile, []byte("hellooooo"))
			Expect(err).ToNot(BeNil())
		})

	})
})
