import {
  fireEvent,
  getByText,
  render,
  RenderResult,
  waitFor,
  waitForElementToBeRemoved,
} from '@testing-library/react';
import * as React from 'react';
import { labels } from '../constants';
import { MoreOptionsMenu, MoreOptionsMenuItem } from '../MoreOptionsMenu';

describe('MoreOptionsMenu', () => {
  let options: MoreOptionsMenuItem[];
  beforeEach(() => {
    options = [
      {
        label: 'Option 1',
        onClick: jest.fn(),
      },
      {
        label: 'Option 2',
        onClick: jest.fn(),
      },
    ];
  });

  const renderMenu = () => render(<MoreOptionsMenu options={options} />);

  const getMenu = async (renderResult: RenderResult) => {
    const { getByLabelText } = renderResult;
    const buttonEl = await waitFor(() => getByLabelText(labels.moreOptionsButton));
    fireEvent.click(buttonEl);
    return waitFor(() => getByLabelText(labels.moreOptionsMenu));
  };

  it('shows menu when button is clicked', async () => {
    const menuEl = await getMenu(renderMenu());
    expect(getByText(menuEl, options[0].label)).toBeInTheDocument();
  });

  it('renders element for each option', async () => {
    const result = renderMenu();
    const menuEl = await getMenu(result);
    expect(getByText(menuEl, options[0].label)).toBeInTheDocument();
    expect(getByText(menuEl, options[1].label)).toBeInTheDocument();
  });

  it('calls handler for item when clicked', async () => {
    const result = renderMenu();
    const menuEl = await getMenu(result);
    const itemEl = getByText(menuEl, options[0].label);
    fireEvent.click(itemEl);
    expect(options[0].onClick).toHaveBeenCalled();
  });

  it('hides menu when item is selected', async () => {
    const result = renderMenu();
    const menuEl = await getMenu(result);
    expect(getByText(menuEl, options[0].label)).toBeInTheDocument();

    const itemEl = getByText(menuEl, options[0].label);
    fireEvent.click(itemEl);
    await waitForElementToBeRemoved(menuEl);
    expect(getByText(menuEl, options[0].label)).not.toBeInTheDocument();
  });

  it('hides menu on escape', async () => {
    const result = renderMenu();
    const menuEl = await getMenu(result);
    expect(getByText(menuEl, options[0].label)).toBeInTheDocument();

    fireEvent.keyDown(menuEl, { key: 'Escape', code: 'Escape' });
    await waitForElementToBeRemoved(menuEl);
    expect(getByText(menuEl, options[0].label)).not.toBeInTheDocument();
  });
});
