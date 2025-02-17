import { Amplify } from 'aws-amplify';
import { generateClient } from 'aws-amplify/api';
import config from './aws-exports';

Amplify.configure(config as any);

export const client = generateClient();
