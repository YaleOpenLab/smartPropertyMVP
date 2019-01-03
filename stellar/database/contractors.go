package database

import (
	utils "github.com/YaleOpenLab/smartPropertyMVP/stellar/utils"
)

func NewContractor(uname string, pwd string, Name string, Address string, Description string) (Entity, error) {
	newContractor, err := NewEntity(uname, pwd, Name, Address, Description, "contractor")
	if err != nil {
		return newContractor, err
	}

	// insert the contractor into the database
	err = InsertEntity(newContractor)
	if err != nil {
		return newContractor, err
	}

	return newContractor, err
}

func (contractor *Entity) ProposeContract(panelSize string, totalValue int, location string, years int, metadata string, recIndex int, projectIndex int) (Project, error) {
	var pc Project
	var err error

	// for this, create a new  contract and store in the contracts db. Wea re sorting
	// by stage, so it shouldn't matter a whole lot
	indexCheck, err := RetrieveAllProjects()
	if err != nil {
		return pc, err
	}
	pc.Params.Index = len(indexCheck) + 1
	pc.Params.PanelSize = panelSize
	pc.Params.TotalValue = totalValue
	pc.Params.Location = location
	pc.Params.Years = years
	pc.Params.Metadata = metadata
	pc.Params.DateInitiated = utils.Timestamp()
	iRecipient, err := RetrieveRecipient(recIndex)
	if err != nil {
		return pc, err
	}
	pc.Params.ProjectRecipient = iRecipient
	pc.Stage = 2 // 2 since we need to filter this out while retrieving the propsoed contracts
	pc.Contractor = *contractor
	// instead of storing in this proposedcontracts slice, store it as a project, but not a contract and retrieve by stage
	err = pc.Save()
	// don't insert the project since the contractor's projects are not final
	return pc, err
}
