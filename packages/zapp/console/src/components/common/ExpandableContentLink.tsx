import { Button, Link } from '@material-ui/core';
import * as React from 'react';

interface ExpandableContentLinkProps {
  /** Set to true to allow hiding the content after it is initially expanded */
  allowCollapse?: boolean;
  /** Use a button instead of a link */
  button?: boolean;
  /** String to show when content is collapsed */
  collapsedText: string;
  /** optional class that will be applied to the wrapping element of the expandable link */
  className?: string;
  /** Optional string to show when content is expanded (only shown if allowCollapse is true). Defaults to "Hide" */
  expandedText?: string;
  /** Optional extra props to pass to the toggle component. Useful, for instance, to add icons to the button. */
  // tslint:disable-next-line:no-any
  extraProps?: { [k: string]: any };
  /** optional class that will be applied to the toggle element */
  toggleClassName?: string;
  /** A callback to notify that the content was expanded */
  onExpand?: () => void;
  /** A callback for rendering the content section */
  renderContent: () => JSX.Element;
}

/**
 * Renders a link which will expand into content when clicked. The content can optionally be collapsed by
 * clicking the link again.
 */
export class ExpandableContentLink extends React.Component<ExpandableContentLinkProps> {
  public static defaultProps = {
    allowCollapse: true,
    button: false,
    expandedText: 'Hide',
  };

  state = {
    showContent: false,
  };

  private renderToggleComponent = () => {
    const { button, collapsedText, expandedText, toggleClassName } = this.props;
    const text = this.state.showContent ? expandedText : collapsedText;
    if (button) {
      return (
        <Button
          variant="outlined"
          className={toggleClassName}
          onClick={this.toggleContent}
          size="small"
          {...this.props.extraProps}
        >
          {text}
        </Button>
      );
    }

    return (
      <Link
        className={toggleClassName}
        component="button"
        variant="body1"
        onClick={this.toggleContent}
      >
        {text}
      </Link>
    );
  };

  private renderExpanded = () => {
    const content = this.props.renderContent();
    return this.props.allowCollapse ? (
      <div className={this.props.className}>
        {this.renderToggleComponent()}
        {content}
      </div>
    ) : (
      content
    );
  };

  private renderCollapsed = () => {
    return <div className={this.props.className}>{this.renderToggleComponent()}</div>;
  };

  private toggleContent = () => {
    const { onExpand } = this.props;
    const { showContent } = this.state;
    this.setState({ showContent: !showContent });
    if (!showContent && onExpand) {
      onExpand();
    }
  };

  public render() {
    return this.state.showContent ? this.renderExpanded() : this.renderCollapsed();
  }
}
