package asn

func IsPrivate(v uint32) bool {
	return (v >= 64512 && v <= 65534) || (v >= 4200000000 && v <= 4294967294)
}

func IsReserved(v uint32) bool {
	return v == 0 || v == 23456 || (v >= 64496 && v <= 64511) || v == 65535 ||
		(v >= 65536 && v <= 65551) || v == 4294967295
}
