import Card from '@material-ui/core/Card';
import CardActionArea from '@material-ui/core/CardActionArea';
import CardContent from '@material-ui/core/CardContent';
import { makeStyles, Theme } from '@material-ui/core/styles';
import Typography from '@material-ui/core/Typography';
import classnames from 'classnames';
import { useCommonStyles } from 'components/common/styles';
import * as React from 'react';
import { Link } from 'react-router-dom';

const useStyles = makeStyles((theme: Theme) => ({
    card: {
        marginBottom: theme.spacing(1)
    },
    link: {
        height: '100%',
        padding: theme.spacing(2)
    }
}));

interface ProjectSectionCardProps {
    description: string;
    label: string;
    link: string;
}

export const ProjectSectionCard: React.FC<ProjectSectionCardProps> = props => {
    const commonStyles = useCommonStyles();
    const styles = useStyles();
    return (
        <Card className={styles.card}>
            <CardActionArea>
                <Link
                    className={classnames(
                        commonStyles.linkUnstyled,
                        styles.link
                    )}
                    to={props.link}
                >
                    <CardContent>
                        <Typography variant="h5">{props.label}</Typography>
                        <div>{props.description}</div>
                    </CardContent>
                </Link>
            </CardActionArea>
        </Card>
    );
};
