import React from 'react';
import {createRoot} from 'react-dom/client';
import {StatusPage} from './StatusPanel.mjs';
import './styles.css';
createRoot(document.getElementById('root')!).render(<StatusPage details={window.location.pathname==='/status/details'}/>);
