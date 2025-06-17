package models

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/OpenListTeam/OpenList-workers/pkg/utils/random"
)

const (
	GENERAL = iota
	GUEST   // only one exists
	ADMIN
)

const StaticHashSalt = "https://github.com/alist-org/alist"

type User struct {
	ID         int       `json:"id" db:"id"`
	Username   string    `json:"username" db:"username" binding:"required"`
	PwdHash    string    `json:"-" db:"pwd_hash"`
	PwdTS      int64     `json:"-" db:"pwd_ts"`
	Salt       string    `json:"-" db:"salt"`
	Password   string    `json:"password,omitempty" db:"-"` // 仅用于输入，不存储
	Role       int       `json:"role" db:"role"`
	Permission int32     `json:"permission" db:"permission"`
	BasePath   string    `json:"base_path" db:"base_path"`
	Disabled   bool      `json:"disabled" db:"disabled"`
	OtpSecret  string    `json:"-" db:"otp_secret"`
	SsoID      string    `json:"sso_id" db:"sso_id"`
	Authn      string    `json:"-" db:"authn"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

func (u *User) IsGuest() bool {
	return u.Role == GUEST
}

func (u *User) IsAdmin() bool {
	return u.Role == ADMIN
}

func (u *User) ValidateRawPassword(password string) error {
	return u.ValidatePwdStaticHash(StaticHash(password))
}

func (u *User) ValidatePwdStaticHash(pwdStaticHash string) error {
	if pwdStaticHash == "" {
		return fmt.Errorf("empty password")
	}
	if u.PwdHash != HashPwd(pwdStaticHash, u.Salt) {
		return fmt.Errorf("wrong password")
	}
	return nil
}

func (u *User) SetPassword(pwd string) *User {
	u.Salt = random.String(16)
	u.PwdHash = TwoHashPwd(pwd, u.Salt)
	u.PwdTS = time.Now().Unix()
	return u
}

func (u *User) CanSeeHides() bool {
	return u.Permission&1 == 1
}

func (u *User) CanAccessWithoutPassword() bool {
	return (u.Permission>>1)&1 == 1
}

func (u *User) CanAddOfflineDownloadTasks() bool {
	return (u.Permission>>2)&1 == 1
}

func (u *User) CanWrite() bool {
	return (u.Permission>>3)&1 == 1
}

func (u *User) CanRename() bool {
	return (u.Permission>>4)&1 == 1
}

func (u *User) CanMove() bool {
	return (u.Permission>>5)&1 == 1
}

func (u *User) CanCopy() bool {
	return (u.Permission>>6)&1 == 1
}

func (u *User) CanRemove() bool {
	return (u.Permission>>7)&1 == 1
}

func (u *User) CanWebdavRead() bool {
	return (u.Permission>>8)&1 == 1
}

func (u *User) CanWebdavManage() bool {
	return (u.Permission>>9)&1 == 1
}

func (u *User) CanFTPAccess() bool {
	return (u.Permission>>10)&1 == 1
}

func (u *User) CanFTPManage() bool {
	return (u.Permission>>11)&1 == 1
}

func (u *User) CanReadArchives() bool {
	return (u.Permission>>12)&1 == 1
}

func (u *User) CanDecompress() bool {
	return (u.Permission>>13)&1 == 1
}

func (u *User) WebAuthnID() []byte {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, uint64(u.ID))
	return bs
}

func (u *User) WebAuthnName() string {
	return u.Username
}

func (u *User) WebAuthnDisplayName() string {
	return u.Username
}

func (u *User) WebAuthnIcon() string {
	return "https://cdn.oplist.org/gh/OpenListTeam/Logo@main/OpenList.svg"
}

func (u *User) WebAuthnCredentials() []interface{} {
	var res []interface{}
	if u.Authn != "" {
		if err := json.Unmarshal([]byte(u.Authn), &res); err != nil {
			return nil
		}
	}
	return res
}

func StaticHash(password string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s-%s", password, StaticHashSalt)))
	return fmt.Sprintf("%x", h)
}

func HashPwd(static string, salt string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s-%s", static, salt)))
	return fmt.Sprintf("%x", h)
}

func TwoHashPwd(password string, salt string) string {
	return HashPwd(StaticHash(password), salt)
}

 