import axios from 'axios';
import { Admin } from 'flyteidl';

import { Identifier } from 'models/Common/types';
import { WorkflowClosure } from 'models/Workflow';

import { NotFoundError } from 'errors';
import { getAdminEntity } from '../AdminEntity';

jest.mock('axios');
const request = axios.request as jest.Mock;

describe('getAdminEntity', () => {
    const messageType = Admin.Workflow;
    let path: string;
    let response: Admin.Workflow;

    beforeEach(() => {
        path = '/workflows/someId';
        response = {
            closure: {} as WorkflowClosure,
            id: {} as Identifier
        };
        request.mockResolvedValue({ data: response });
    });

    it('sets correct method and headers', () => {
        getAdminEntity({ path, messageType });
        expect(request).toHaveBeenCalledWith(
            expect.objectContaining({
                headers: { Accept: 'application/octet-stream' },
                method: 'get',
                responseType: 'arraybuffer'
            })
        );
    });

    it('passes through `params` values', () => {
        const params = { someQuery: 'abcd' };
        getAdminEntity({ path, messageType }, { params });
        expect(request).toHaveBeenCalledWith(
            expect.objectContaining({
                params: expect.objectContaining({ ...params })
            })
        );
    });

    it('decodes protobuf responses using the provided message class', async () => {
        const decodedValue = { ...response };
        const messageClass = {
            decode: jest.fn(() => decodedValue)
        };
        const returnedValue = await getAdminEntity({
            path,
            messageType: messageClass
        });
        expect(messageClass.decode).toHaveBeenCalled();
        expect(returnedValue).toBe(decodedValue);
    });

    it('returns a transformed response', async () => {
        const transform = jest.fn();
        await getAdminEntity({ messageType, path, transform });
        expect(transform).toHaveBeenCalled();
    });

    describe('with data', () => {
        const data = { someProp: 'someValue' };
        it('passes through `data` values', () => {
            getAdminEntity({ path, messageType }, { data });
            expect(request).toHaveBeenCalledWith(
                expect.objectContaining({
                    data
                })
            );
        });

        it('sets correct content type', () => {
            getAdminEntity({ path, messageType }, { data });
            expect(request).toHaveBeenCalledWith(
                expect.objectContaining({
                    headers: expect.objectContaining({
                        'Content-Type': 'application/octet-stream'
                    })
                })
            );
        });
    });

    describe('on error', () => {
        it('Returns a NotFoundError for 404 responses', async () => {
            request.mockRejectedValue({ response: { status: 404 } });
            await expect(getAdminEntity({ path, messageType })).rejects.toEqual(
                new NotFoundError(path)
            );
        });
    });
});
