package types

// ValidateBasic is used for validating the packet
func (p FungibleTokenTransferPacketData) ValidateBasic() error {

	// TODO: Validate the packet data

	return nil
}

// GetBytes is a helper for serialising
func (p FungibleTokenTransferPacketData) GetBytes() ([]byte, error) {
	var modulePacket AssetPacketData

	modulePacket.Packet = &AssetPacketData_FungibleTokenTransferPacket{&p}

	return modulePacket.Marshal()
}
