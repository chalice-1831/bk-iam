/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云-权限中心(BlueKing-IAM) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package dao

//go:generate mockgen -source=$GOFILE -destination=./mock/$GOFILE -package=mock

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"iam/pkg/database"
)

// AuthExpression ...
type AuthExpression struct {
	PK         int64  `db:"pk"`
	Expression string `db:"expression"`
	Signature  string `db:"signature"`
}

// Expression ...
type Expression struct {
	PK         int64  `db:"pk"`
	Type       int64  `db:"type"`
	Expression string `db:"expression"`
	Signature  string `db:"signature"`
}

// ExpressionManager ...
type ExpressionManager interface {
	// for auth

	ListAuthByPKs(pks []int64) ([]AuthExpression, error)

	// for saas

	ListDistinctBySignaturesType(signatures []string, _type int64) ([]Expression, error)

	BulkCreateWithTx(tx *sqlx.Tx, expressions []Expression) (int64, error) // 返回批量创建的last id
	BulkUpdateWithTx(tx *sqlx.Tx, expressions []Expression) error
	BulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) (int64, error)
}

type expressionManager struct {
	DB *sqlx.DB
}

// NewExpressionManager ...
func NewExpressionManager() ExpressionManager {
	return &expressionManager{
		DB: database.GetDefaultDBClient().DB,
	}
}

// ListAuthByPKs ...
func (m *expressionManager) ListAuthByPKs(pks []int64) (expressions []AuthExpression, err error) {
	if len(pks) == 0 {
		return
	}
	err = m.selectAuthByPKs(&expressions, pks)
	if errors.Is(err, sql.ErrNoRows) {
		return expressions, nil
	}
	return
}

// ListDistinctBySignaturesType List distinct expressions by signatures and type
func (m *expressionManager) ListDistinctBySignaturesType(
	signatures []string, _type int64,
) (expressions []Expression, err error) {
	if len(signatures) == 0 {
		return
	}
	err = m.selectBySignaturesType(&expressions, signatures, _type)
	if errors.Is(err, sql.ErrNoRows) {
		return expressions, nil
	}
	return
}

// BulkCreateWithTx ...
func (m *expressionManager) BulkCreateWithTx(tx *sqlx.Tx, expressions []Expression) (int64, error) {
	if len(expressions) == 0 {
		return 0, nil
	}
	return m.bulkInsertWithTx(tx, expressions)
}

// BulkUpdateWithTx ...
func (m *expressionManager) BulkUpdateWithTx(tx *sqlx.Tx, expressions []Expression) error {
	if len(expressions) == 0 {
		return nil
	}
	return m.bulkUpdateWithTx(tx, expressions)
}

// BulkDeleteByPKsWithTx ...
func (m *expressionManager) BulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) (int64, error) {
	if len(pks) == 0 {
		return 0, nil
	}
	return m.bulkDeleteByPKsWithTx(tx, pks)
}

func (m *expressionManager) selectAuthByPKs(expressions *[]AuthExpression, pks []int64) error {
	query := `SELECT
		pk,
		expression,
		signature
		FROM expression
		WHERE pk IN (?)`
	return database.SqlxSelect(m.DB, expressions, query, pks)
}

func (m *expressionManager) selectBySignaturesType(expressions *[]Expression, signatures []string, _type int64) error {
	query := `SELECT
		pk,
		type,
		expression,
		signature
		FROM expression
		WHERE pk IN (
			SELECT
			MIN(pk)
			FROM expression
			WHERE signature IN (?)
			AND type = ?
			GROUP BY signature
		)`
	return database.SqlxSelect(m.DB, expressions, query, signatures, _type)
}

func (m *expressionManager) bulkInsertWithTx(tx *sqlx.Tx, expressions []Expression) (int64, error) {
	sql := `INSERT INTO expression (
		type,
		expression,
		signature
	) VALUES (
		:type,
		:expression,
		:signature)`
	return database.SqlxBulkInsertReturnIDWithTx(tx, sql, expressions)
}

func (m *expressionManager) bulkUpdateWithTx(tx *sqlx.Tx, expressions []Expression) error {
	sql := `UPDATE expression SET
		expression=:expression,
		signature=:signature
		WHERE pk=:pk`
	return database.SqlxBulkUpdateWithTx(tx, sql, expressions)
}

func (m *expressionManager) bulkDeleteByPKsWithTx(tx *sqlx.Tx, pks []int64) (int64, error) {
	sql := `DELETE FROM expression WHERE pk IN (?)`
	return database.SqlxDeleteReturnRowsWithTx(tx, sql, pks)
}
