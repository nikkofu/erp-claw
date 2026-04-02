package supplychain

type BackofficeOverview struct {
	TenantID   string
	Payable    PayableOverview
	Receivable ReceivableOverview
	Sales      SalesOverview
}

type PayableOverview struct {
	OpenCount int
}

type ReceivableOverview struct {
	OpenCount int
}

type SalesOverview struct {
	DraftCount   int
	ShippedCount int
	TotalCount   int
}
