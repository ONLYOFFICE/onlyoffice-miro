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

import React, {
  forwardRef,
  FormEvent,
  useState,
  useEffect,
  useRef,
} from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import {
  validateAddress,
  validateShortText,
  normalizeAddress,
} from '@utils/validator';

import Button from '@components/Button';
import FormInput from '@components/Input';

import { Banner } from '@features/settings/components/Banner';

import useSettingsStore from '@features/settings/stores/useSettingsStore';
import useApplicationStore from '@stores/useApplicationStore';
import useEmitterStore from '@stores/useEmitterStore';

import { sanitizeUrl, sanitizeFormInput } from '@utils/sanitizer';

import '@features/settings/components/form.css';

interface FormProps extends React.HTMLAttributes<HTMLDivElement> {}

export const Form = forwardRef<HTMLDivElement, FormProps>(
  ({ className, children, ...props }, ref) => {
    const [addressError, setAddressError] = useState('');
    const [secretError, setSecretError] = useState('');
    const [headerError, setHeaderError] = useState('');
    const [submitting, setSubmitting] = useState(false);
    const { t } = useTranslation();
    const navigate = useNavigate();
    const hasSaved = useRef(false);
    const {
      address,
      header,
      secret,
      loading,
      demo,
      demoStarted,
      setAddress,
      setHeader,
      setSecret,
      setDemo,
      saveSettings,
      saveOriginalValues,
      revertToOriginalValues,
      hasUnsavedChanges,
    } = useSettingsStore();
    const { refreshAuthorization } = useApplicationStore();
    const { emitRefreshDocuments } = useEmitterStore();

    const isDemoExpired = demoStarted
      ? (() => {
          const startTime = new Date(demoStarted).getTime();
          const expiryDays = parseInt(
            import.meta.env.VITE_ASC_DEMO_EXPIRATION_DAYS || '30',
            10
          );
          const expiryTime = startTime + expiryDays * 24 * 60 * 60 * 1000;
          const currentTime = Date.now();
          return currentTime > expiryTime;
        })()
      : false;

    useEffect(() => {
      saveOriginalValues();
    }, [saveOriginalValues]);

    useEffect(() => {
      return () => {
        if (!hasSaved.current && hasUnsavedChanges()) {
          revertToOriginalValues();
        }
      };
    }, [hasUnsavedChanges, revertToOriginalValues]);

    useEffect(() => {
      if (demo && !isDemoExpired) {
        setAddress('');
        setHeader('');
        setSecret('');
        setAddressError('');
        setHeaderError('');
        setSecretError('');
      }
    }, [demo, isDemoExpired, setAddress, setHeader, setSecret]);

    const hasInputs =
      address.trim() !== '' || header.trim() !== '' || secret.trim() !== '';

    const fieldsRequired = isDemoExpired || !demo;

    const validateAddressField = (value: string): string => {
      if (!fieldsRequired) return '';
      return validateAddress(value)
        ? ''
        : t('features.settings.form.errors.address_required');
    };

    const validateHeaderField = (value: string): string => {
      if (!fieldsRequired) return '';
      return validateShortText(value)
        ? ''
        : t('features.settings.form.errors.header_required');
    };

    const validateSecretField = (value: string): string => {
      if (!fieldsRequired) return '';
      return validateShortText(value)
        ? ''
        : t('features.settings.form.errors.secret_required');
    };

    const addressErr = validateAddressField(address);
    const headerErr = validateHeaderField(header);
    const secretErr = validateSecretField(secret);

    const hasValidationErrors = !!(addressErr || headerErr || secretErr);

    const saveDisabled =
      loading ||
      submitting ||
      (fieldsRequired && (!hasInputs || hasValidationErrors)) ||
      (!fieldsRequired && hasInputs);

    const demoCheckboxDisabled = loading || submitting || isDemoExpired;

    const validateForm = (): boolean => {
      if (!fieldsRequired) return true;

      const currentAddressErr = validateAddressField(address);
      const currentHeaderErr = validateHeaderField(header);
      const currentSecretErr = validateSecretField(secret);

      setAddressError(currentAddressErr);
      setHeaderError(currentHeaderErr);
      setSecretError(currentSecretErr);

      return !currentAddressErr && !currentHeaderErr && !currentSecretErr;
    };

    const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      setSubmitting(true);
      try {
        if (!fieldsRequired || validateForm()) {
          try {
            await saveSettings();
            hasSaved.current = true;
            await emitRefreshDocuments();
            await refreshAuthorization();
            navigate('/');
          } catch (err: unknown) {
            if (err && typeof err === 'object' && 'message' in err) {
              miro.board.notifications.showError(
                t((err as { message: string }).message)
              );
            }
          }
        }
      } finally {
        setSubmitting(false);
      }
    };

    return (
      <div ref={ref} className={`form ${className || ''}`} {...props}>
        <div className="form__content">
          <p className="form__description">
            {t('features.settings.form.description')}
          </p>
          <form
            onSubmit={handleSubmit}
            className="form__fields"
            autoComplete="off"
          >
            <div className="form__field">
              <FormInput
                label={t('features.settings.form.address')}
                name="address"
                type="text"
                value={address}
                error={addressError}
                disabled={loading || submitting || (demo && !isDemoExpired)}
                onChange={(e) => {
                  const { value } = e.target;
                  const sanitizedValue = sanitizeUrl(value);
                  setAddress(sanitizedValue);
                }}
                onBlur={(e) => {
                  const sanitized = sanitizeUrl(e.target.value);
                  const normalized = normalizeAddress(sanitized);
                  if (normalized !== e.target.value) {
                    setAddress(normalized);
                  }
                  
                  setAddressError(validateAddressField(normalized));
                }}
                required={fieldsRequired}
                autoComplete="off"
              />
            </div>
            <div className="form__field">
              <FormInput
                label={t('features.settings.form.secret')}
                name="secret"
                type="password"
                value={secret}
                error={secretError}
                disabled={loading || submitting || (demo && !isDemoExpired)}
                onChange={(e) => {
                  const { value } = e.target;
                  const sanitizedValue = sanitizeFormInput(value);
                  setSecret(sanitizedValue);
                }}
                onBlur={(e) => {
                  setSecretError(validateSecretField(e.target.value));
                }}
                required={fieldsRequired}
                autoComplete="off"
              />
            </div>
            <div className="form__field">
              <FormInput
                label={t('features.settings.form.header')}
                name="header"
                type="text"
                value={header}
                error={headerError}
                disabled={loading || submitting || (demo && !isDemoExpired)}
                onChange={(e) => {
                  const { value } = e.target;
                  const sanitizedValue = sanitizeFormInput(value);
                  setHeader(sanitizedValue);
                }}
                onBlur={(e) => {
                  setHeaderError(validateHeaderField(e.target.value));
                }}
                required={fieldsRequired}
                autoComplete="off"
              />
            </div>

            {isDemoExpired && (
              <div className="form__checkbox-container form__checkbox-container__expired">
                <p className="form__checkbox-description form__checkbox-description__expired">
                  {t('features.settings.form.demo.expired')}
                </p>
              </div>
            )}

            <Banner />

            {!isDemoExpired && (
              <div className="form__checkbox-container">
                <label className="form__checkbox-label" htmlFor="demo-checkbox">
                  <input
                    id="demo-checkbox"
                    type="checkbox"
                    className="checkbox form__checkbox"
                    style={{ margin: '0', marginRight: '8px' }}
                    checked={demo}
                    disabled={demoCheckboxDisabled}
                    onChange={() => {
                      setDemo(!demo);
                      if (!demo) {
                        setAddress('');
                        setHeader('');
                        setSecret('');
                        setAddressError('');
                        setHeaderError('');
                        setSecretError('');
                      }
                    }}
                  />
                  <span className="form__checkbox-text">
                    {t('features.settings.form.demo.title')}
                  </span>
                </label>
                <p className="form__checkbox-description">
                  {!demoStarted && t('features.settings.form.demo.description')}
                  {demoStarted &&
                    t('features.settings.form.demo.available_until', {
                      date: new Date(
                        new Date(demoStarted).getTime() +
                          parseInt(
                            import.meta.env.VITE_ASC_DEMO_EXPIRATION_DAYS ||
                              '30',
                            10
                          ) *
                            86400000
                      )
                        .toLocaleDateString('en-GB')
                        .replace(/\//g, '.'),
                    })}
                </p>
              </div>
            )}

            <div className="form__button-container">
              <Button
                type="submit"
                name={t('features.settings.form.save')}
                variant="primary"
                disabled={saveDisabled}
                className="form__save-button"
                title={t('features.settings.form.save')}
              />
            </div>
          </form>
        </div>
      </div>
    );
  }
);

Form.displayName = 'Form';

export default Form;
