package vmservice

type CreateVmParams struct {
	Namespace  string
	VmName     string
	VmPassword string
	dnsHost    string
	VmSSHPort  int32
	VmImage    string
	VmDiskNum  string
	UserID     uint
}
