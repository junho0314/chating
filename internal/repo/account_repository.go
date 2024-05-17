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
 * Authors: Jungmin Eum
 */

package repo

import (
	"chating_service/internal/constants"
	"chating_service/internal/db"
	"chating_service/internal/model"

	"time"

	"github.com/rs/zerolog/log"
)

// IsUserIdInDatabase returns true if the userId is in database
func IsUserIdInDatabase(dbCtx *db.DbCtx, userId string) (bool, error) {
	selectSQL := `
		SELECT EXISTS(
			SELECT 1 
			FROM ACCOUNT 
			WHERE user_id = ?
		)
	`
	var exists bool
	err := dbCtx.DB.QueryRow(selectSQL, userId).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func CreateAccount(dbCtx *db.DbCtx, account *model.NewAccountForm, companyId int64) error {
	insertSQL := `
		INSERT INTO ACCOUNT
			(
				id, 
				user_id, 
				password, 
				created_at, 
				created_by
			) 
		VALUES 
			(?,?,?,current_timestamp(),?)
	`

	_, err := dbCtx.DB.ExecContext(dbCtx.Ctx, insertSQL,
		account.Id,
		companyId,
		account.UserId,
		account.Password,
		false,
		constants.AccountDefaultStatus,
		constants.AuthDefaultStatus,
		time.Now().Format("20060102"),
		constants.ServerName,
	)
	if err != nil {
		return err
	}

	log.Info().Msgf("InsertAccount:: last Insert Id : %d ", account.Id)
	return nil
}

func UpdateAccountPassword(dbCtx *db.DbCtx, newEncryptedPassword string, accountId int64) error {
	updateSQL := `
		UPDATE ACCOUNT
			SET password = ?,
			    change_password_latest_date = current_timestamp(),
			    updated_at = current_timestamp(),
			    updated_by = ?
		WHERE id = ?
	`
	_, err := dbCtx.DB.ExecContext(dbCtx.Ctx, updateSQL,
		newEncryptedPassword,
		constants.ServerName,
		accountId)
	if err != nil {
		return err
	}

	return nil
}

func UpsertPushToken(dbCtx *db.DbCtx, accountId int64, pushToken string) error {
	upsertSQL := `
			INSERT INTO PUSH_ENDPOINT
				(
				account_id,
				push_token,
				created_at,
				created_by
				)
			VALUES (?,?,current_timestamp(),?)
			ON DUPLICATE KEY UPDATE 
				push_token=?,
				updated_at=current_timestamp(),
				updated_by=?
		`
	_, err := dbCtx.DB.ExecContext(dbCtx.Ctx, upsertSQL,
		accountId,
		pushToken,
		constants.ServerName,
		pushToken,
		constants.ServerName)
	if err != nil {
		return err
	}

	return nil
}

func GetUserAccountByAccountId(dbCtx *db.DbCtx, accountId int64) (model.Account, error) {
	account := model.Account{}
	selectQuery := `
		SELECT id, 
		       user_id, 
		       password, 
		       is_used, 
        FROM ACCOUNT 
        WHERE id=?
	`
	stmt, err := dbCtx.CreatePrepareStmt(selectQuery)
	if err != nil {
		return account, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(accountId).Scan(
		&account.Id,
		&account.UserId,
		&account.Password,
		&account.IsUsed,
	)
	if err != nil {
		log.Error().Msg(
			"GetUserAccount:: error while fetching user account" + err.Error())
		return account, err
	}

	return account, nil
}

func GetUserAccount(dbCtx *db.DbCtx, userId string) (model.Account, error) {
	account := model.Account{}
	selectQuery := `
		SELECT id, 
		       user_id,
		       password, 
		       is_used
        FROM ACCOUNT 
        WHERE user_id=?
	`
	stmt, err := dbCtx.CreatePrepareStmt(selectQuery)
	if err != nil {
		return account, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(userId).Scan(
		&account.Id,
		&account.UserId,
		&account.Password,
		&account.IsUsed,
	)
	if err != nil {
		log.Error().Msg(
			"GetUserAccount:: error while fetching user account" + err.Error())
		return account, err
	}
	return account, nil
}
