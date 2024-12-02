package client

// CS 161 Project 2

// Only the following imports are allowed! ANY additional imports
// may break the autograder!
// - bytes
// - encoding/hex
// - encoding/json
// - errors
// - fmt
// - github.com/cs161-staff/project2-userlib
// - github.com/google/uuid
// - strconv
// - strings

import (
	"encoding/json"
	"strconv"

	userlib "github.com/cs161-staff/project2-userlib"
	"github.com/google/uuid"

	// hex.EncodeToString(...) is useful for converting []byte to string

	// Useful for string manipulation

	// Useful for formatting strings (e.g. `fmt.Sprintf`).

	// Useful for creating new error messages to return using errors.New("...")
	"errors"

	// Optional.
	_ "strconv"
)

const MAX_NODE_CONTENT_SIZE = 256

// Linked-List FileBody definition
type FileBody struct {
	Data uuid.UUID
	Next uuid.UUID
}

// File Head definition
type FileHead struct {
	Alpha  uint16
	Beta   uint16
	First  uuid.UUID
	Last   uuid.UUID
	Length int
}

type Intermediate struct {
	InfoPointer uuid.UUID
	EncKey      []byte
	MacKey      []byte
	IsOwner     bool
}

// This is the type definition for the User struct.
// A Go struct is like a Python or Java class - it can have attributes
// (e.g. like the Username attribute) and methods (e.g. like the StoreFile method below).
type User struct {
	Username   string
	SignKey    userlib.DSSignKey
	PrivateKey userlib.PKEDecKey
	SharedByMe []uuid.UUID

	// You can add other attributes here if you want! But note that in order for attributes to
	// be included when this struct is serialized to/from JSON, they must be capitalized.
	// On the flipside, if you have an attribute that you want to be able to access from
	// this struct's methods, but you DON'T want that value to be included in the serialized value
	// of this struct that's stored in datastore, then you can use a "private" variable (e.g. one that
	// begins with a lowercase letter).
}

type ShareHub struct {
	FileHeadUUID uuid.UUID
	MacKey       []byte
	EncKey       []byte
}

type Share struct {
	ShareHubUUID uuid.UUID
	MacKey       []byte
	EncKey       []byte
}

type ShareList struct {
	List []SharePair
}

type SharePair struct {
	Filename    string
	Recipient   string
	Sharestruct Share
}

/**
 * Fast Hash username given by user
 * Get user uuid from datastore
 * Chceck if username is empty or if username already exists: return error
 * initialize user struct and return address to user struct
 */
func InitUser(username string, password string) (userdataptr *User, err error) {
	userPointer := GenerateUserUUID(username)
	_, userExists := userlib.DatastoreGet(uuid.UUID(userPointer))
	if len(username) <= 0 || userExists {
		return nil, errors.New("username already exists or no username entered	")
	}
	rsaPublic, rsaPrivate, rsaErr := userlib.PKEKeyGen()
	if rsaErr != nil {
		return nil, rsaErr
	}
	keystoreErr := userlib.KeystoreSet(username+"RSA", rsaPublic)
	if keystoreErr != nil {
		return nil, keystoreErr
	}
	signPrivate, signPublic, signErr := userlib.DSKeyGen()
	if signErr != nil {
		return nil, signErr
	}
	keystoreErr = userlib.KeystoreSet(username+"Verify", signPublic)
	if keystoreErr != nil {
		return nil, keystoreErr
	}
	user := User{username, signPrivate, rsaPrivate, []uuid.UUID{}}
	marshaledUser, marshalErr := json.Marshal(user)
	if marshalErr != nil {
		return nil, marshalErr
	}
	encryptionKey, verificationKey := GenerateUserKeys(username, password)
	encryptedUser, etmErr := EncryptThenMac(marshaledUser, encryptionKey, verificationKey, userlib.RandomBytes(16))
	if etmErr != nil {
		return nil, etmErr
	}
	userlib.DatastoreSet(userPointer, encryptedUser)

	shareListPointer, byteErr := uuid.FromBytes(userlib.Hash([]byte(username + "shareList"))[:16])
	if byteErr != nil {
		return nil, byteErr
	}
	shareListMarshaled, marshalErr := json.Marshal(ShareList{})
	shareListEncrypted, etsErr := EncryptThenSign(signPrivate, username, shareListMarshaled)
	if etsErr != nil {
		return nil, etsErr
	}
	userlib.DatastoreSet(shareListPointer, shareListEncrypted)

	return &user, nil
}

/**
 * Generate UUID, user struct decryption key with and verification key. User
 * can login iff they can verify the HMAC and decrypt their user struct, else
 * throw error.
 */
func GetUser(username string, password string) (userdataptr *User, err error) {
	var userPointer = GenerateUserUUID(username)
	var decryptionKey, verificationKey = GenerateUserKeys(username, password)
	// Retrieve HMACed and encrypted user struct from Datastore
	var userStruct, datastoreGetErr = userlib.DatastoreGet(userPointer)
	if !datastoreGetErr {
		return nil, errors.New("Datastore retrieval error.")
	}
	// check HMAC then decrypt user struct
	userStruct, mtdErr := MacThenDecrypt(userStruct, decryptionKey, verificationKey)
	if mtdErr != nil {
		return nil, mtdErr
	}
	// Unmarshal decrypted struct into object
	var userData User
	var unmarshalErr = json.Unmarshal(userStruct, &userData)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &userData, nil
}

func (userdata *User) StoreFile(filename string, content []byte) (err error) {
	// Create intermediate and upload
	intermediatePointer, byteErr := uuid.FromBytes(userlib.Hash([]byte(userdata.Username + filename))[:16])
	if byteErr != nil {
		return byteErr
	}
	fileHeadPointer, byteErr := uuid.FromBytes(userlib.RandomBytes(16))
	if byteErr != nil {
		return byteErr
	}
	fileHeadEncKey, fileHeadMacKey := userlib.RandomBytes(16), userlib.RandomBytes(16)
	intermediate := Intermediate{fileHeadPointer, fileHeadEncKey, fileHeadMacKey, true}
	intermediateMarshaled, marshalErr := json.Marshal(intermediate)
	if marshalErr != nil {
		return marshalErr
	}
	intermediateEncrypted, etsErr := EncryptThenSign(userdata.SignKey, userdata.Username, intermediateMarshaled)
	if etsErr != nil {
		return etsErr
	}
	userlib.DatastoreSet(intermediatePointer, intermediateEncrypted)

	// Create file head and upload
	// FIXME: alpha, beta only 8 bits => collisions imminent
	alpha, beta := uint16(userlib.RandomBytes(16)[0]), uint16(userlib.RandomBytes(16)[0])
	firstBodyUUID := uuid.New()
	fileHead := FileHead{uint16(alpha), uint16(beta), firstBodyUUID, firstBodyUUID, 1}
	fileHeadMarshaled, marshalErr := json.Marshal(fileHead)
	if marshalErr != nil {
		return marshalErr
	}
	fileHeadEncrypted, etmErr := EncryptThenMac(fileHeadMarshaled, fileHeadEncKey, fileHeadMacKey, userlib.RandomBytes(16))
	if etmErr != nil {
		return etmErr
	}
	userlib.DatastoreSet(fileHeadPointer, fileHeadEncrypted)

	// Create and upload first node
	firstDataUUID := uuid.New()
	dataEncKey, dataMacKey, nodeEncKey, nodeMacKey, gbkErr := GenerateBodyKeys(alpha, beta, fileHead.Length-1)
	if gbkErr != nil {
		return gbkErr
	}
	firstBody := FileBody{firstDataUUID, uuid.Nil}
	firstBodyMarshaled, marshalErr := json.Marshal(firstBody)
	if marshalErr != nil {
		return marshalErr
	}
	firstBodyEncrypted, etmErr := EncryptThenMac(firstBodyMarshaled, nodeEncKey, nodeMacKey, userlib.RandomBytes(16))
	if etmErr != nil {
		return etmErr
	}
	userlib.DatastoreSet(firstBodyUUID, firstBodyEncrypted)

	// Add content
	var firstData []byte
	if len(content) > MAX_NODE_CONTENT_SIZE {
		firstData = content[:MAX_NODE_CONTENT_SIZE]
		appendErr := userdata.AppendToFile(filename, content[MAX_NODE_CONTENT_SIZE:])
		if appendErr != nil {
			return appendErr
		}
	} else {
		firstData = content
	}
	firstDataEncrypted, etmErr := EncryptThenMac(firstData, dataEncKey, dataMacKey, userlib.RandomBytes(16))
	if etmErr != nil {
		return etmErr
	}
	userlib.DatastoreSet(firstDataUUID, firstDataEncrypted)

	return nil
}

func (userdata *User) AppendToFile(filename string, content []byte) error {
	// Get File head and Alpha/Beta
	fileHead, fileUUID, err := GetFileHead(filename, userdata.Username, userdata.PrivateKey)
	if err != nil {
		return err
	}
	alpha := fileHead.Alpha
	beta := fileHead.Beta

	//Loop for multiple nodes
	for len(content) > 0 {
		// Determine the size of content for this node.
		sizeToWrite := len(content)
		if sizeToWrite > MAX_NODE_CONTENT_SIZE {
			sizeToWrite = MAX_NODE_CONTENT_SIZE
		}

		// Create a chunk of content to write to the node.
		contentChunk := content[:sizeToWrite]

		// Generate new keys based on the new node index (fileHead.Length)
		// Generate Encryption and Mac key for data and node using alpha and beta
		dataEncKey, dataMacKey, nodeEncKey, nodeMacKey, gbkErr := GenerateBodyKeys(alpha, beta, fileHead.Length)
		if gbkErr != nil {
			return gbkErr
		}

		// Encrypt and upload Content to datastore
		cipherData, err := EncryptThenMac(contentChunk, dataEncKey, dataMacKey, userlib.RandomBytes(16))
		if err != nil {
			return err
		}
		cipherDataUUID := uuid.New()
		userlib.DatastoreSet(cipherDataUUID, cipherData)

		// Marshal Node, encrypt and upload Node to Datastore and get UUID
		var newLastNode FileBody
		newLastNode.Data = cipherDataUUID
		newLastNode.Next = uuid.Nil
		MarshalledNewLastNode, err := json.Marshal(newLastNode)
		if err != nil {
			return err
		}
		cipherNode, err := EncryptThenMac(MarshalledNewLastNode, nodeEncKey, nodeMacKey, userlib.RandomBytes(16))
		if err != nil {
			return err
		}
		CipherNodeUUID := uuid.New()
		userlib.DatastoreSet(CipherNodeUUID, cipherNode)

		// Create keys from beta to Decrypt Last Node and change next of last Node to new Last Node uuid(cipherNodeUUID),
		// then decrypt and reupload to datastore
		// userlib.Hash([]byte(strconv.Itoa(int(beta))))[:16]
		OldLastEncKey, err := userlib.HashKDF(userlib.Hash([]byte(strconv.Itoa(int(beta))))[:16], []byte(strconv.Itoa(fileHead.Length-1)+"encrypt"))
		OldLastEncKey = OldLastEncKey[:16]
		if err != nil {
			return err
		}
		OldLastMacKey, err := userlib.HashKDF(userlib.Hash([]byte(strconv.Itoa(int(beta))))[:16], []byte(strconv.Itoa(fileHead.Length-1)+"mac"))
		OldLastMacKey = OldLastMacKey[:16]
		if err != nil {
			return err
		}
		OldLastNodeUUID := fileHead.Last
		cipherOldLastNode, ok := userlib.DatastoreGet(OldLastNodeUUID)
		if !ok {
			return errors.New("could not retrieve last node of file from datastore")
		}
		MarshalledOldLastNode, err := MacThenDecrypt(cipherOldLastNode, OldLastEncKey, OldLastMacKey)
		if err != nil {
			return err
		}
		var oldLastNode FileBody
		err = json.Unmarshal(MarshalledOldLastNode, &oldLastNode)
		if err != nil {
			return err
		}
		oldLastNode.Next = CipherNodeUUID
		MarshalledOldLastNode, err = json.Marshal(oldLastNode)
		if err != nil {
			return err
		}
		cipherOldLastNode, err = EncryptThenMac(MarshalledOldLastNode, OldLastEncKey, OldLastMacKey, userlib.RandomBytes(16))
		if err != nil {
			return err
		}
		userlib.DatastoreSet(OldLastNodeUUID, cipherOldLastNode)

		// lastly change filehead.last to newLastNodeUUID(cipherUUID) and update the length
		// and prepare content for next iteration
		content = content[sizeToWrite:]
		fileHead.Last = CipherNodeUUID
		fileHead.Length += 1
	}

	// Finally encrypt filehead with symetric encryption and upload to datastore
	marshalledFileHead, err := json.Marshal(fileHead)
	if err != nil {
		return err
	}
	intermediate, err := GetIntermediate(filename, userdata.Username, userdata.PrivateKey)
	if err != nil {
		return err
	}
	var encFileHead []byte
	if intermediate.IsOwner {
		encFileHead, err = EncryptThenMac(marshalledFileHead, intermediate.EncKey, intermediate.MacKey, userlib.RandomBytes(16))
		if err != nil {
			return err
		}
	} else {
		shareHub, gshErr := GetShareHub(intermediate.InfoPointer, intermediate.EncKey, intermediate.MacKey)
		if gshErr != nil {
			return gshErr
		}
		encFileHead, err = EncryptThenMac(marshalledFileHead, shareHub.EncKey, shareHub.MacKey, userlib.RandomBytes(16))
	}
	userlib.DatastoreSet(fileUUID, encFileHead)
	return nil
}

func (userdata *User) LoadFile(filename string) (content []byte, err error) {
	fileHead, _, err := GetFileHead(filename, userdata.Username, userdata.PrivateKey)
	if err != nil {
		return nil, err
	}
	alpha, beta, currNodeUUID, n, i := fileHead.Alpha, fileHead.Beta, fileHead.First, fileHead.Length, 0
	content = []byte{}

	for i < n {
		dataEncKey, dataMacKey, nodeEncKey, nodeMacKey, gbkErr := GenerateBodyKeys(alpha, beta, i)
		if gbkErr != nil {
			return nil, gbkErr
		}
		// get node
		currNodeRaw, nodeExists := userlib.DatastoreGet(currNodeUUID)
		if !nodeExists {
			return nil, errors.New("Node " + strconv.Itoa(i) + " does not exist")
		}
		currNodeRaw, mtdErr := MacThenDecrypt(currNodeRaw, nodeEncKey, nodeMacKey)
		if mtdErr != nil {
			return nil, mtdErr
		}
		var currNode FileBody
		unmarshalErr := json.Unmarshal(currNodeRaw, &currNode)
		if unmarshalErr != nil {
			return nil, unmarshalErr
		}
		// get content
		currDataRaw, dataExists := userlib.DatastoreGet(currNode.Data)
		if !dataExists {
			return nil, errors.New("Data " + strconv.Itoa(i) + " does not exist")
		}
		currData, mtdErr := MacThenDecrypt(currDataRaw, dataEncKey, dataMacKey)
		if mtdErr != nil {
			return nil, mtdErr
		}
		content = append(content, currData...)
		currNodeUUID = currNode.Next
		i += 1
	}
	return content, nil
}

func (userdata *User) CreateInvitation(filename string, recipientUsername string) (
	invitationPtr uuid.UUID, err error) {
	intermediate, intermediateErr := GetIntermediate(filename, userdata.Username, userdata.PrivateKey)
	if intermediateErr != nil {
		return uuid.Nil, intermediateErr
	}

	var share Share
	if intermediate.IsOwner {
		shareListPointer, byteErr := uuid.FromBytes(userlib.Hash([]byte(userdata.Username + "shareList"))[:16])
		if byteErr != nil {
			return uuid.Nil, byteErr
		}
		shareListRaw, ok := userlib.DatastoreGet(shareListPointer)
		if !ok {
			return uuid.Nil, errors.New("unable to retrieve shareList")
		}
		shareListRaw, vtdErr := VerifyThenDecrypt(userdata.Username, userdata.PrivateKey, shareListRaw)
		if vtdErr != nil {
			return uuid.Nil, vtdErr
		}
		var shareList ShareList
		unmarshalErr := json.Unmarshal(shareListRaw, &shareList)
		if unmarshalErr != nil {
			return uuid.Nil, unmarshalErr
		}
		//share pair stuff used to be here
		shareHub := ShareHub{intermediate.InfoPointer, intermediate.MacKey, intermediate.EncKey}
		shareHubEncKey, shareHubMacKey := userlib.RandomBytes(16), userlib.RandomBytes(16)
		shareHubMarshaled, marshalErr := json.Marshal(shareHub)
		if marshalErr != nil {
			return uuid.Nil, marshalErr
		}
		shareHubEncrypted, mtdErr := EncryptThenMac(shareHubMarshaled, shareHubEncKey, shareHubMacKey, userlib.RandomBytes(16))
		if mtdErr != nil {
			return uuid.Nil, mtdErr
		}
		shareHubPointer, byteErr := uuid.FromBytes(userlib.Hash([]byte(userdata.Username + recipientUsername + filename))[:16])
		if byteErr != nil {
			return uuid.Nil, byteErr
		}
		userlib.DatastoreSet(shareHubPointer, shareHubEncrypted)
		share = Share{shareHubPointer, shareHubMacKey, shareHubEncKey}
		// added the shair pair stuff here so that sharestruct can be included
		sharePair := SharePair{filename, recipientUsername, share}
		shareList.List = append(shareList.List, sharePair)
		shareListMarshaled, marshalErr := json.Marshal(shareList)
		if marshalErr != nil {
			return uuid.Nil, marshalErr
		}
		shareListEncrypted, etsErr := EncryptThenSign(userdata.SignKey, userdata.Username, shareListMarshaled)
		if etsErr != nil {
			return uuid.Nil, etsErr
		}
		userlib.DatastoreSet(shareListPointer, shareListEncrypted)
	} else {
		// Do we need this in sharepair??
		share = Share{intermediate.InfoPointer, intermediate.MacKey, intermediate.EncKey}
	}
	shareRaw, marshalErr := json.Marshal(share)
	if marshalErr != nil {
		return uuid.Nil, marshalErr
	}
	shareEncrypted, etsErr := EncryptThenSign(userdata.SignKey, recipientUsername, shareRaw)
	if etsErr != nil {
		return uuid.Nil, etsErr
	}
	sharePointer, byteErr := uuid.FromBytes(userlib.RandomBytes(16))
	if byteErr != nil {
		return uuid.Nil, byteErr
	}
	userlib.DatastoreSet(sharePointer, shareEncrypted)
	return sharePointer, nil
}

func (userdata *User) AcceptInvitation(senderUsername string, invitationPtr uuid.UUID, filename string) error {
	shareRaw, shareExists := userlib.DatastoreGet(invitationPtr)
	if !shareExists {
		return errors.New("share does not exist")
	}
	shareRaw, vtdErr := VerifyThenDecrypt(senderUsername, userdata.PrivateKey, shareRaw)
	if vtdErr != nil {
		return vtdErr
	}
	var share Share
	marshalErr := json.Unmarshal(shareRaw, &share)
	if marshalErr != nil {
		return marshalErr
	}
	intermediate := Intermediate{share.ShareHubUUID, share.EncKey, share.MacKey, false}
	intermediateMarshaled, marshalErr := json.Marshal(intermediate)
	if marshalErr != nil {
		return marshalErr
	}
	intermediateEncrypted, etsErr := EncryptThenSign(userdata.SignKey, userdata.Username, intermediateMarshaled)
	if etsErr != nil {
		return etsErr
	}
	intermediatePointer, byteErr := uuid.FromBytes(userlib.Hash([]byte(userdata.Username + filename))[:16])
	if byteErr != nil {
		return byteErr
	}
	userlib.DatastoreSet(intermediatePointer, intermediateEncrypted)
	return nil
}

func (userdata *User) RevokeAccess(filename string, recipientUsername string) error {
	//Reencrypt file and replace orignal file at intermediate
	LoadedFile, err := userdata.LoadFile(filename)
	if err != nil {
		return err
	}

	//Go to revoked sharehub and replace with garbage and remove revoked user from sharelist
	RevokedShareHubUUID, err := uuid.FromBytes(userlib.Hash([]byte(userdata.Username + recipientUsername + filename))[:16])
	if err != nil {
		return err
	}
	garbage, err := json.Marshal(ShareHub{})
	if err != nil {
		return err
	}
	userlib.DatastoreSet(RevokedShareHubUUID, garbage)

	userdata.StoreFile(filename, LoadedFile)

	//Get sharelist
	shareListPointer, byteErr := uuid.FromBytes(userlib.Hash([]byte(userdata.Username + "shareList"))[:16])
	if byteErr != nil {
		return byteErr
	}
	shareListRaw, ok := userlib.DatastoreGet(shareListPointer)
	if !ok {
		return errors.New("unable to retrieve shareList")
	}
	shareListRaw, vtdErr := VerifyThenDecrypt(userdata.Username, userdata.PrivateKey, shareListRaw)
	if vtdErr != nil {
		return vtdErr
	}
	var shareList ShareList
	unmarshalErr := json.Unmarshal(shareListRaw, &shareList)
	if unmarshalErr != nil {
		return unmarshalErr
	}

	intermediate, err := GetIntermediate(filename, userdata.Username, userdata.PrivateKey)
	if err != nil {
		return nil
	}
	//Update all non revoked sharehubs for this file by iterating through share list and checking sharepair for the particular file
	for i := 0; i < len(shareList.List); i++ {
		if shareList.List[i].Filename == filename {
			if shareList.List[i].Recipient == recipientUsername { // remove or replace revoked shairpair with an empty shairpair
				// TODO: fix remove?
				shareList.List[i] = SharePair{}
			} else { //update all other sharehubs
				//get Sharehub
				ShareHubUUID, err := uuid.FromBytes(userlib.Hash([]byte(userdata.Username + shareList.List[i].Recipient + filename))[:16])
				if err != nil {
					return err
				}

				ShareHubEncKey, ShareHubMacKey := shareList.List[i].Sharestruct.EncKey, shareList.List[i].Sharestruct.MacKey

				//update sharehub
				shareHub := ShareHub{intermediate.InfoPointer, intermediate.MacKey, intermediate.EncKey}
				shareHubMarshaled, marshalErr := json.Marshal(shareHub)
				if marshalErr != nil {
					return marshalErr
				}
				shareHubEncrypted, mtdErr := EncryptThenMac(shareHubMarshaled, ShareHubEncKey, ShareHubMacKey, userlib.RandomBytes(16))
				if mtdErr != nil {
					return mtdErr
				}
				userlib.DatastoreSet(ShareHubUUID, shareHubEncrypted)
			}
		}
	}

	// Store Sharelist
	shareListRaw, err = json.Marshal(shareList)
	if err != nil {
		return err
	}
	sharelistEncrypted, err := EncryptThenSign(userdata.SignKey, recipientUsername, shareListRaw)
	if err != nil {
		return err
	}
	userlib.DatastoreSet(shareListPointer, sharelistEncrypted)

	return nil
}

func GenerateUserUUID(username string) (generatedUUID uuid.UUID) {
	var ret, err = uuid.FromBytes(userlib.Hash([]byte(username))[:16])
	if err != nil {
		panic(err)
	}
	return ret
}

func GenerateUserKeys(username string, password string) (encryptionKey []byte, verificationKey []byte) {
	var usernameBytes = []byte(username)
	var usernameXored0x5c []byte
	var usernameXored0x36 []byte
	for i := 0; i < len(usernameBytes); i++ {
		usernameXored0x5c = append(usernameXored0x5c, usernameBytes[i]^0x5c)
		usernameXored0x36 = append(usernameXored0x36, usernameBytes[i]^0x36)
	}
	encryptionKey = userlib.Argon2Key([]byte(password), usernameXored0x5c, 16)
	verificationKey = userlib.Argon2Key([]byte(password), usernameXored0x36, 16)
	return encryptionKey, verificationKey
}

func GenerateBodyKeys(alpha uint16, beta uint16, i int) (dataEncKey []byte, dataMacKey []byte, nodeEncKey []byte, nodeMacKey []byte, err error) {
	hashKeyAlpha := userlib.Hash([]byte(strconv.Itoa(int(alpha))))[:16]
	hashKeyBeta := userlib.Hash([]byte(strconv.Itoa(int(beta))))[:16]
	dataEncKey, err = userlib.HashKDF(hashKeyAlpha, []byte(strconv.Itoa(i)+"encrypt"))
	dataEncKey = dataEncKey[:16]
	if err != nil {
		return nil, nil, nil, nil, err
	}
	dataMacKey, err = userlib.HashKDF(hashKeyAlpha, []byte(strconv.Itoa(i)+"mac"))
	dataMacKey = dataMacKey[:16]
	if err != nil {
		return nil, nil, nil, nil, err
	}
	nodeEncKey, err = userlib.HashKDF(hashKeyBeta, []byte(strconv.Itoa(i)+"encrypt"))
	nodeEncKey = nodeEncKey[:16]
	if err != nil {
		return nil, nil, nil, nil, err
	}
	nodeMacKey, err = userlib.HashKDF(hashKeyBeta, []byte(strconv.Itoa(i)+"mac"))
	nodeMacKey = nodeMacKey[:16]
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return dataEncKey, dataMacKey, nodeEncKey, nodeMacKey, nil
}

/**
 * Compute HMAC on the file and evaluate with the given HMAC. Return plaintext if they match, return error
 * if they don't
 */
func MacThenDecrypt(file []byte, decryptionKey []byte, verificationKey []byte) (plaintext []byte, err error) {
	var originalHmac = file[:64]
	var fileContents = file[64:]
	var recomputedHmac, hmacErr = userlib.HMACEval(verificationKey, fileContents)
	if hmacErr != nil {
		return nil, hmacErr
	}
	if !userlib.HMACEqual(originalHmac, recomputedHmac) {
		return nil, errors.New("HMACs were not equal.")
	}
	plaintext = userlib.SymDec(decryptionKey, fileContents)
	return plaintext, nil
}

/**
 * Encrypt the file and then MAC the ciphertext. Return the HMAC concatenated with the ciphertext
 */
func EncryptThenMac(plaintext []byte, encryptionKey []byte, verificationKey []byte, iv []byte) (encryptedAndHmacedFile []byte, err error) {
	var ciphertext = userlib.SymEnc(encryptionKey, iv, plaintext)
	var hmac, hmacErr = userlib.HMACEval(verificationKey, ciphertext)
	if hmacErr != nil {
		return nil, hmacErr
	}
	encryptedAndHmacedFile = append(hmac, ciphertext...)
	return encryptedAndHmacedFile, nil
}

/**
 * Verify digital signature using public verification key and decrypt with RSA private Key
 * I am returning the unmarshalled verified and decrypted []byte that needs to be marshalled
 */
func VerifyThenDecrypt(SenderUsername string, receiverPrivateKey userlib.PKEDecKey, dataToDecrypt []byte) (verifiedAndDecryptedBytes []byte, err error) {
	VerifyKey, ok := userlib.KeystoreGet(SenderUsername + "Verify")
	if !ok {
		return nil, errors.New("verification key not found in keystore")
	}
	var signature = dataToDecrypt[:256]
	var content = dataToDecrypt[256:]
	VerifyErr := userlib.DSVerify(VerifyKey, content, signature)
	if VerifyErr != nil {
		return nil, VerifyErr
	}
	encryptedKey := content[:256]
	content = content[256:]
	decryptionKey, decErr := userlib.PKEDec(receiverPrivateKey, encryptedKey)
	if decErr != nil {
		return nil, decErr
	}
	verifiedAndDecryptedBytes = userlib.SymDec(decryptionKey, content)
	return verifiedAndDecryptedBytes, nil
}

func EncryptThenSign(senderSignKey userlib.DSSignKey, receiverUsername string, dataToEncrypt []byte) (encryptedAndSignedBytes []byte, err error) {
	encKey := userlib.RandomBytes(16)
	cipherText := userlib.SymEnc(encKey, userlib.RandomBytes(16), dataToEncrypt)
	publicEncKey, ok := userlib.KeystoreGet(receiverUsername + "RSA")
	if !ok {
		return nil, errors.New("receiver's public Si not found in keystore")
	}
	encryptedKey, encryptionErr := userlib.PKEEnc(publicEncKey, encKey)
	if encryptionErr != nil {
		return nil, encryptionErr
	}
	keyAndCiphertext := append(encryptedKey, cipherText...)
	signature, signErr := userlib.DSSign(senderSignKey, keyAndCiphertext)
	if signErr != nil {
		return nil, signErr
	}
	// This return value consists as follows:
	// [0 : 255] -- Signature
	// [256 : 256 + 256 = 272] -- Encrypted encryption key
	// [256 + 256: ] -- Ciphertext
	encryptedAndSignedBytes = append(signature, keyAndCiphertext...)
	return encryptedAndSignedBytes, nil
}

func GetFileHead(fileName string, username string, RSAPrivateKey userlib.PrivateKeyType) (fileHead FileHead, fileHeadUUID uuid.UUID, err error) {

	// Get Intermediate
	intermediate, err := GetIntermediate(fileName, username, RSAPrivateKey)
	if err != nil {
		return FileHead{}, uuid.Nil, err
	}

	if intermediate.IsOwner {
		// Get File head for owner
		fileHeadToBeMaccedThenDecrypted, ok := userlib.DatastoreGet(intermediate.InfoPointer)
		if !ok {
			return FileHead{}, uuid.Nil, errors.New("file doesnt exist at given filename")
		}
		fileHeadMarshaled, err := MacThenDecrypt(fileHeadToBeMaccedThenDecrypted, intermediate.EncKey, intermediate.MacKey)
		if err != nil {
			return FileHead{}, uuid.Nil, err
		}
		err = json.Unmarshal(fileHeadMarshaled, &fileHead)
		if err != nil {
			return FileHead{}, uuid.Nil, err
		}
		return fileHead, intermediate.InfoPointer, nil
	} else {
		shareHub, gshErr := GetShareHub(intermediate.InfoPointer, intermediate.EncKey, intermediate.MacKey)
		if gshErr != nil {
			return FileHead{}, uuid.Nil, gshErr
		}

		// get file head from sharehub
		fileHeadToBeMaccedThenDecrypted, ok := userlib.DatastoreGet(shareHub.FileHeadUUID)
		if !ok {
			return FileHead{}, uuid.Nil, errors.New("file doesnt exist at given filename")
		}
		fileHeadMarshaled, err := MacThenDecrypt(fileHeadToBeMaccedThenDecrypted, shareHub.EncKey, shareHub.MacKey)
		if err != nil {
			return FileHead{}, uuid.Nil, err
		}
		err = json.Unmarshal(fileHeadMarshaled, &fileHead)
		if err != nil {
			return FileHead{}, uuid.Nil, err
		}
		return fileHead, shareHub.FileHeadUUID, nil
	}
}

func GetShareHub(pointer userlib.UUID, encKey []byte, macKey []byte) (shareHub ShareHub, err error) {
	shareHubToBeMaccedThenDecrypted, ok := userlib.DatastoreGet(pointer)
	if !ok {
		return ShareHub{}, errors.New("sharehub doesnt exist at given infopointer")
	}
	shareHubMarshaled, err := MacThenDecrypt(shareHubToBeMaccedThenDecrypted, encKey, macKey)
	if err != nil {
		return ShareHub{}, err
	}
	err = json.Unmarshal(shareHubMarshaled, &shareHub)
	if err != nil {
		return ShareHub{}, err
	}
	return shareHub, nil
}

func GetIntermediate(fileName string, username string, RSAPrivateKey userlib.PrivateKeyType) (intermediate Intermediate, err error) {
	fileIntermediateUUID, err := uuid.FromBytes(userlib.Hash([]byte(username + fileName))[:16])
	if err != nil {
		return Intermediate{}, err
	}
	marshalledEncryptedIntermediate, ok := userlib.DatastoreGet(fileIntermediateUUID)
	if !ok {
		return Intermediate{}, errors.New("cannot retrieve intermediate from datastore")
	}
	marshalledDecryptedIntermediate, err := VerifyThenDecrypt(username, RSAPrivateKey, marshalledEncryptedIntermediate)
	if err != nil {
		return Intermediate{}, err
	}
	err = json.Unmarshal(marshalledDecryptedIntermediate, &intermediate)
	if err != nil {
		return Intermediate{}, err
	}
	return intermediate, nil
}
