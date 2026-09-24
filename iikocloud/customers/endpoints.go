// Package customers paths. Every path is a constant here; none appears in a method body.

package customers

const (
	addCustomerCard                   = "/api/1/loyalty/iiko/customer/card/add"
	addCustomerCategory               = "/api/1/loyalty/iiko/customer_category/add"
	addCustomerProgram                = "/api/1/loyalty/iiko/customer/program/add"
	calculateLoyalty                  = "/api/1/loyalty/iiko/calculate"
	cancelHoldCustomerWallet          = "/api/1/loyalty/iiko/customer/wallet/cancel_hold"
	chargeoffCustomerWallet           = "/api/1/loyalty/iiko/customer/wallet/chargeoff"
	checkSMSSendingPossibility        = "/api/1/loyalty/iiko/check_sms_sending_possibility"
	checkSMSStatus                    = "/api/1/loyalty/iiko/check_sms_status"
	createOrUpdateCustomer            = "/api/1/loyalty/iiko/customer/create_or_update"
	deleteCustomers                   = "/api/1/loyalty/iiko/delete_customers"
	getCounters                       = "/api/1/loyalty/iiko/get_counters"
	getCouponsBySeries                = "/api/1/loyalty/iiko/coupons/by_series"
	getCouponsInfo                    = "/api/1/loyalty/iiko/coupons/info"
	getCouponsSeries                  = "/api/1/loyalty/iiko/coupons/series"
	getCustomerCategory               = "/api/1/loyalty/iiko/customer_category"
	getCustomerInfo                   = "/api/1/loyalty/iiko/customer/info"
	getCustomerTransactionsByDate     = "/api/1/loyalty/iiko/customer/transactions/by_date"
	getCustomerTransactionsByRevision = "/api/1/loyalty/iiko/customer/transactions/by_revision"
	getManualCondition                = "/api/1/loyalty/iiko/manual_condition"
	getProgram                        = "/api/1/loyalty/iiko/program"
	holdCustomerWallet                = "/api/1/loyalty/iiko/customer/wallet/hold"
	removeCustomerCard                = "/api/1/loyalty/iiko/customer/card/remove"
	removeCustomerCategory            = "/api/1/loyalty/iiko/customer_category/remove"
	restoreCustomers                  = "/api/1/loyalty/iiko/restore_customers"
	sendEmailMessage                  = "/api/1/loyalty/iiko/message/send_email"
	sendSMSMessage                    = "/api/1/loyalty/iiko/message/send_sms"
	topupCustomerWallet               = "/api/1/loyalty/iiko/customer/wallet/topup"
)
