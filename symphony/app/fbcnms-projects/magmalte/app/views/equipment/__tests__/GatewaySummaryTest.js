/**
 * Copyright 2004-present Facebook. All Rights Reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree.
 *
 * @flow strict-local
 * @format
 */
import 'jest-dom/extend-expect';
import GatewaySummary from '../GatewaySummary';
import MagmaAPIBindings from '@fbcnms/magma-api';
import React from 'react';
import axiosMock from 'axios';
import {MemoryRouter, Route} from 'react-router-dom';
import {cleanup, render, wait} from '@testing-library/react';
import type {lte_gateway} from '@fbcnms/magma-api';

jest.mock('axios');
jest.mock('@fbcnms/magma-api');
jest.mock('@fbcnms/ui/hooks/useSnackbar');

afterEach(cleanup);

const mockGatewaySt: lte_gateway = {
  cellular: {
    epc: {
      ip_block: '',
      nat_enabled: true,
    },
    ran: {
      pci: 260,
      transmit_enabled: true,
    },
  },
  connected_enodeb_serials: [],
  description: 'mpk_dogfooding',
  device: {
    hardware_id: 'e059637f-cd55-4109-816c-ce6ebc69020d',
    key: {
      key: '',
      key_type: 'SOFTWARE_ECDSA_SHA256',
    },
  },
  id: 'mpk_dogfooding_magma_1',
  magmad: {
    autoupgrade_enabled: true,
    autoupgrade_poll_interval: 301,
    checkin_interval: 60,
    checkin_timeout: 20,
  },
  name: 'team pod',
  status: {
    hardware_id: 'e059637f-cd55-4109-816c-ce6ebc69020d',
    platform_info: {
      packages: [
        {
          name: 'magma',
          version: '1.1.0-1590005479-e6e781a9',
        },
      ],
    },
  },
  tier: 'default',
};

describe('<GatewaySummary />', () => {
  beforeEach(() => {
    // eslint-disable-next-line max-len
    MagmaAPIBindings.getLteByNetworkIdGatewaysByGatewayId.mockResolvedValue(
      mockGatewaySt,
    );
  });

  afterEach(() => {
    axiosMock.get.mockClear();
  });

  const Wrapper = () => (
    <MemoryRouter
      initialEntries={['/nms/mynetwork/gateway/mygateway']}
      initialIndex={0}>
      <Route
        path="/nms/:networkId/gateway/:gatewayId"
        component={GatewaySummary}
      />
    </MemoryRouter>
  );

  it('renders', async () => {
    const {container} = render(<Wrapper />);
    await wait();
    expect(
      MagmaAPIBindings.getLteByNetworkIdGatewaysByGatewayId,
    ).toHaveBeenCalledTimes(1);
    expect(container).toHaveTextContent('mpk_dogfooding');
    expect(container).toHaveTextContent('1.1.0-1590005479-e6e781a9');
    expect(container).toHaveTextContent('e059637f-cd55-4109-816c-ce6ebc69020d');
    expect(container).toHaveTextContent('mpk_dogfooding_magma_1');
  });
});
