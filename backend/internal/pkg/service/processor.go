/**
 *
 * (c) Copyright Ascensio System SIA 2025
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package service

type StorageProcessor[ID comparable, T any, R any] interface {
	TableName() string
	BuildSelectQuery(id ID) (query string, args []any, scanner func(R) (T, error))
	BuildInsertQuery(id ID, component T) (query string, args []any)
	BuildUpdateQuery(id ID, component T) (query string, args []any)
	BuildDeleteQuery(id ID) (query string, args []any)
}
