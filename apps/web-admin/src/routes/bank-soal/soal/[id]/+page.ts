import type { PageLoad } from './$types';
import { detailRouteQuestionId } from '../../_components/soal-workspace.navigation';

export const load: PageLoad = ({ params }) => ({ questionId: detailRouteQuestionId(params.id) });
