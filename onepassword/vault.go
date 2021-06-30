package onepassword

import (
	"encoding/json"
)

const VaultResource = "vault"

type Vault struct {
	UUID string
	Name string
}

func (o *OnePassClient) ReadVault(id string) (*Vault, error) {
	vault := &Vault{}
	args := []string{opPasswordGet, VaultResource, id}
	res, err := o.runCmd(args...)
	if err != nil {
		return nil, prettyError(args, res, err)
	}
	if err = json.Unmarshal(res, vault); err != nil {
		return nil, err
	}
	return vault, nil
}

func (o *OnePassClient) CreateVault(v *Vault) (*Vault, error) {
	args := []string{opPasswordCreate, VaultResource, v.Name}
	res, err := o.runCmd(args...)
	if err != nil {
		return nil, prettyError(args, res, err)
	}
	if err = json.Unmarshal(res, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (o *OnePassClient) DeleteVault(id string) error {
	return o.Delete(VaultResource, id)
}
