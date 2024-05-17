/**
 *
 * MOTREXEV CONFIDENTIAL
 * _____________________________________________________________________
 *
 * [2023] - [2050] MOTREXEV
 *  All Rights Reserved.
 *
 * NOTICE:  All information contained herein is, and remains
 * the property of MOTREXEV and its suppliers,
 * if any.  The intellectual and technical concepts contained
 * herein are proprietary to MOTREXEV and its suppliers and
 * may be covered by Korea and Foreign Patents,
 * patents in process, and are protected by trade secret or copyright law.
 * Dissemination of this information or reproduction of this material
 * is strictly forbidden unless prior written permission is obtained
 * from MOTREXEV.
 *
 * Authors: Kumar Ankur, Jungmin Eum, Junho Hong
 */

package constants

const (
	Success = 0
	Error   = -1

	// 계정
	InvalidUserId             = 1001
	InvalidPassword           = 1002
	InvalidAccountStatus      = 1003
	InvalidAuthStatus         = 1004
	InvalidCredentials        = 1005
	SameAsCurrentPassword     = 1006
	InvalidCompanyEmailDomain = 1007
	EmailDuplicate            = 1008

	ExistItem          = 2001
	ThereIsNoData      = 2002
	BadRequest         = 2003
	InvalidInputData   = 2004
	InvalidFilters     = 2005
	NotExistItem       = 2007
	CheckRequiredItems = 2006
	ExceedMaxCount     = 2008
	ExceedMaxLength    = 2009
	UnremovableItem    = 2010
	UnupdatableItem    = 2011

	// ota
	InvalidProductId  = 2101
	ExceedReleaseNote = 2102
	ExceedComment     = 2103
	ExceedVersion     = 2104

	ErrorTariffPlanNotExist = 3005

	InvalidFileSize      = 4001
	InvalidFileExtension = 4002
	InvalidFileCheckSum  = 4003

	// 제품
	InvalidProductManufacturer = 7001
	InvalidSoftwareProfile     = 7002

	EmailSendingFailed  = 9001
	ServerInternalError = 9999
)
