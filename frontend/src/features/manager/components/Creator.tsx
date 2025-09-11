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

import React, { forwardRef, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import { FileInfo } from '@lib/types';

import Button from '@components/Button';
import FormInput from '@components/Input';
import Label from '@components/Label';
import Select, { SelectOption } from '@components/Select';

import useCreatorStore from '@features/manager/stores/useCreatorStore';
import useEmitterStore from '@stores/useEmitterStore';

import { sanitizeFilename } from '@utils/sanitizer';

import '@features/manager/components/creator.css';

interface CreatorProps extends React.HTMLAttributes<HTMLDivElement> {}

export const Creator = forwardRef<HTMLDivElement, CreatorProps>(
  ({ className, children, ...props }, ref) => {
    const { t } = useTranslation();
    const navigate = useNavigate();
    const {
      selectedName,
      selectedType,
      loading,
      setSelectedName,
      setSelectedType,
      resetSelected,
      createFile,
      getSupportedTypes,
    } = useCreatorStore();
    const { emitDocumentCreated } = useEmitterStore();

    const [nameError, setNameError] = useState<string | undefined>(undefined);

    useEffect(() => {
      setSelectedType(getSupportedTypes()[0]);
    }, [getSupportedTypes, setSelectedType]);

    useEffect(() => {
      if (!selectedName || selectedName.trim() === '')
        setNameError(t('features.manager.creation.errors.name_required'));
      else setNameError(undefined);
    }, [selectedName, t]);

    const typeOptions: SelectOption[] = useMemo(() => {
      const supportedTypes = getSupportedTypes();
      return supportedTypes.map((type) => ({
        value: type,
        label: t(`features.manager.creation.file_types.${type}`),
      }));
    }, [getSupportedTypes, t]);

    const formValid =
      selectedName && selectedName.trim() !== '' && selectedType && !nameError;

    const handleCreateFile = async () => {
      const createdFile = await createFile();
      if (!createdFile) return null;

      emitDocumentCreated({
        id: createdFile.id,
        name: `${selectedName}.${selectedType}`,
        type: selectedType,
        createdAt: createdFile.createdAt,
        modifiedAt: createdFile.modifiedAt,
        links: {
          self: createdFile.links.self,
        },
      } as FileInfo);
      resetSelected();

      return createdFile;
    };

    return (
      <div
        ref={ref}
        className={`creator-container ${className || ''}`}
        {...props}
      >
        <div className="creator-container__main">
          <div className="creator-container_shifted">
            <div className="creator-container__form">
              <div className="creator-container__form-group">
                <Label className="creator-container__label" htmlFor="file-name">
                  {t('features.manager.creation.file_name')}
                  <span className="form-input__label_required">*</span>
                </Label>
                <FormInput
                  id="file-name"
                  className="creator-container__input"
                  value={selectedName}
                  onChange={(e) => {
                    const sanitizedName = sanitizeFilename(e.target.value);
                    setSelectedName(sanitizedName);
                  }}
                  disabled={loading}
                  error={nameError}
                />
              </div>
              <div className="creator-container__form-group">
                <Label className="creator-container__label" htmlFor="file-type">
                  {t('features.manager.creation.file_type')}
                </Label>
                <Select
                  options={typeOptions}
                  value={selectedType}
                  onChange={setSelectedType}
                  className="creator-container__select"
                  disabled={loading}
                />
              </div>
            </div>
          </div>
        </div>
        <div className="creator-container__button-container">
          <Button
            name={
              loading
                ? t('features.manager.creation.creating')
                : t('features.manager.creation.create')
            }
            variant="primary"
            onClick={async () => {
              const result = await handleCreateFile();
              if (result) navigate('/', { state: { isBack: true } });
              else
                await miro.board.notifications.showError(
                  t('features.manager.creation.errors.file_creation_failed')
                );
            }}
            disabled={loading || !formValid}
            style={{
              width: '100%',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}
          />
        </div>
      </div>
    );
  }
);

Creator.displayName = 'Creator';

export default Creator;
