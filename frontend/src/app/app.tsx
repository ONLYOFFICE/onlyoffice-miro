import React, {
  StrictMode,
  Suspense,
  useCallback,
  useEffect,
  useRef,
  useState,
} from 'react';
import { HashRouter, Routes, Route, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { createRoot } from 'react-dom/client';

import { CSSTransition, SwitchTransition } from 'react-transition-group';

import { CenterLayout } from '@components/CenterLayout';
import Spinner from '@components/Spinner';

import CreationPage from '@app/pages/creation';
import Installation from '@features/manager/components/Installation';
import ManagerPage from '@app/pages/manager';
import SettingsPage from '@app/pages/settings';

import useFilesStore from '@features/file/stores/useFileStore';
import useApplicationStore from '@stores/useApplicationStore';
import { EmitterEvents } from '@stores/useEmitterStore';

import '@app/transitions.css';
import '@i18n/config';

const App = () => {
  const nodeRef = useRef(null);
  const location = useLocation();
  const { i18n } = useTranslation();
  const { refreshDocuments } = useFilesStore();
  const { loading, authorized, admin, reloadAuthorization } =
    useApplicationStore();

  const [prevPathname, setPrevPathname] = useState(location.pathname);

  const changeLocale = useCallback(async () => {
    const userInfo = await miro.board.getInfo();
    i18n.changeLanguage(userInfo.locale);
  }, [i18n]);

  useEffect(() => {
    changeLocale().then(reloadAuthorization).then(refreshDocuments);

    miro?.board.events.on(EmitterEvents.REFRESH_DOCUMENTS, refreshDocuments);

    return () => {
      miro?.board.events.off(EmitterEvents.REFRESH_DOCUMENTS, refreshDocuments);
    };
  }, [changeLocale, refreshDocuments, reloadAuthorization]);

  useEffect(() => {
    const isBack = prevPathname.length > location.pathname.length;
    if (!location.state) location.state = { isBack };
    setPrevPathname(location.pathname);
  }, [location, location.pathname, prevPathname.length]);

  if (loading)
    return (
      <CenterLayout style={{ height: '100vh' }}>
        <Spinner size="large" />
      </CenterLayout>
    );

  if (!authorized) return <Installation />;

  return (
    <div className="page-container">
      <Suspense fallback={<Spinner size="large" />}>
        <SwitchTransition mode="out-in">
          <CSSTransition
            key={location.pathname}
            nodeRef={nodeRef}
            timeout={300}
            classNames={location.state?.isBack ? 'page-back' : 'page-forward'}
            unmountOnExit
          >
            <div ref={nodeRef}>
              <Routes location={location}>
                <Route path="/" element={<ManagerPage />} />
                <Route path="/create" element={<CreationPage />} />
                {admin && <Route path="/settings" element={<SettingsPage />} />}
              </Routes>
            </div>
          </CSSTransition>
        </SwitchTransition>
      </Suspense>
    </div>
  );
};

const rootElement = document.getElementById('root');
if (rootElement) {
  createRoot(rootElement).render(
    <StrictMode>
      <HashRouter>
        <App />
      </HashRouter>
    </StrictMode>
  );
}
