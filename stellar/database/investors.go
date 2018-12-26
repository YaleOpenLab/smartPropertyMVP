package database

// contains the WIP Investor struct which will be st ored in a separate bucket
import (
	"encoding/json"
	"log"

	utils "github.com/YaleOpenLab/smartPropertyMVP/stellar/utils"
	"github.com/boltdb/bolt"
	"github.com/stellar/go/build"
)

// the investor struct contains all the investor details such as
// public key, seed (if account is created on the website) and ot her stuff which
// is yet to be decided

// All investors will be referenced by their public key, name is optional (maybe necessary?)
// we need to stil ldecide on identity and stuff and how much we want to track
// people who invest in the schools
type Investor struct {
	VotingBalance int // this will be equal to the amount of stablecoins that the
	// investor possesses
	AmountInvested float64
	// total amount, would be nice to track to contact them,
	// give them some kind of medals or something
	InvestedAssets []Order
	// array of asset codes this user has invested in
	// also I think we need a username + password for logging on to the platform itself
	// linking it here for now
	U User
	// user related functions are called as an instance directly
}

// NewInvestor creates a new investor object when passed the username, password hash,
// name and an option to generate the seed and publicKey. This is done because if
// we decide to allow anonymous investors to invest on our platform, we can easily
// insert their publickey into the system and hten have hanlders for them signing
// transactions
// TODO: add anonymous investor signing handlers
func NewInvestor(uname string, pwd string, Name string) (Investor, error) {
	// call this after the user has failled in username and password. Store hashed password
	// in the database
	var a Investor
	var err error
	a.U, err = NewUser(uname, pwd, Name)
	if err != nil {
		return a, err
	}
	a.AmountInvested = float64(0)
	return a, nil
}

// InsertInvestor inserts a passed Investor object into the database
func InsertInvestor(a Investor) error {
	db, err := OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(InvestorBucket)
		encoded, err := json.Marshal(a)
		if err != nil {
			log.Println("Failed to encode this data into json")
			return err
		}
		return b.Put([]byte(utils.ItoB(a.U.Index)), encoded)
		// but why do we index based on Index?
		// this is because we do want to enumerate through all investors, which can not be done
		// in a name based construction. But this makes search ahrder, since now you
		// all entries to find something as simple as a password. But if this is the
		// only use case that exists, we index by password hash and then get data only
		// when the user requests it. Nice data protection as well
		// TODO: discuss indexing by pwd hash and implications. For small no of entries,
		// we can s till tierate over all the entries.
	})
	return err
}

// RetrieveAllInvestors gets a list of all investor in the database
func RetrieveAllInvestors() ([]Investor, error) {
	// this route is broken becuase it reads through keys sequentially
	// need to see keys until the length of the users database
	var arr []Investor
	temp, err := RetrieveAllUsers()
	if err != nil {
		return arr, err
	}
	limit := len(temp) + 1
	db, err := OpenDB()
	if err != nil {
		return arr, err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(InvestorBucket)
		for i := 1; i < limit; i++ {
			var rInvestor Investor
			x := b.Get(utils.ItoB(i))
			if x == nil {
				// this is where the key does not exist
				continue
			}
			err := json.Unmarshal(x, &rInvestor)
			//if err != nil && rInvestor.Live == false {
			if err != nil {
				// we've reached the end of input, so this is not an error
				// ideal error would be "unexpected JSON input" or something similar
				return nil
			}
			arr = append(arr, rInvestor)
		}
		return nil
	})
	return arr, err
}

// RetrieveInvestor retrieves a particular investor indexed by key from the database
func RetrieveInvestor(key int) (Investor, error) {
	var inv Investor
	db, err := OpenDB()
	if err != nil {
		return inv, err
	}
	defer db.Close()
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(InvestorBucket)
		x := b.Get(utils.ItoB(key))
		if x == nil {
			return nil
		}
		return json.Unmarshal(x, &inv)
	})
	return inv, nil
}

func ValidateInvestor(name string, pwhash string) (Investor, error) {
	var rec Investor
	user, err := ValidateUser(name, pwhash)
	if err != nil {
		return rec, err
	}
	return RetrieveInvestor(user.Index)
}


func (a *Investor) DeductVotingBalance(votes int) error {
	// TODO: we need to update the voting balance often in accordance with the stablecoin
	// balance or a user will have way less votes. This needs an aadditional field
	// in the db to track past balance and then adjust the amoutn of votes he has
	// accordingly
	var err error
	a.VotingBalance -= votes
	err = InsertInvestor(*a)
	if err != nil {
		return err
	}
	return nil
}

func (a *Investor) AddVotingBalance(votes int) error {
	// this funtion is caled when we want to refund the user with the votes once
	// an order has been finalized.
	// TODO: use this
	var err error
	a.VotingBalance += votes
	err = InsertInvestor(*a)
	if err != nil {
		return err
	}
	return nil
}


// TrustAsset creates a trustline from the caller towards the specific asset
// and asset issuer with a _limit_ set on the maximum amount of tokens that can be sent
// through the trust channel. Each trustline costs 0.5XLM.
func (a *Investor) TrustAsset(asset build.Asset, limit string) (string, error) {
	// TRUST is FROM recipient TO issuer
	trustTx, err := build.Transaction(
		build.SourceAccount{a.U.PublicKey},
		build.AutoSequence{SequenceProvider: utils.DefaultTestNetClient},
		build.TestNetwork,
		build.Trust(asset.Code, asset.Issuer, build.Limit(limit)),
	)

	if err != nil {
		return "", err
	}

	trustTxe, err := trustTx.Sign(a.U.Seed)
	if err != nil {
		return "", err
	}

	trustTxeB64, err := trustTxe.Base64()
	if err != nil {
		return "", err
	}

	tx, err := utils.DefaultTestNetClient.SubmitTransaction(trustTxeB64)
	if err != nil {
		return "", err
	}

	log.Println("Trusted asset tx: ", tx.Hash)
	return tx.Hash, nil
}
