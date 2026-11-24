import http from 'k6/http';
import { check, sleep } from 'k6';

const preparacao = __ENV.PREPARACAO === 'true';
const nomes = ['servico-1', 'servico-2', 'servico-3', 'servico-4', 'servico-5'];
const baseURL = __ENV.BASE_URL || 'http://gatus:8080';

export const options = {
  vus: Number(__ENV.VUS || 1),
  duration: __ENV.DURACAO || '2m',
  summaryTrendStats: ['avg', 'p(95)', 'max'],
  thresholds: preparacao ? {} : {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate==0'],
    checks: ['rate==1'],
  },
};

export default function () {
  const resposta = http.get(
    `${baseURL}/api/v1/endpoints/statuses?page=1&pageSize=50`,
    { timeout: '5s', redirects: 0 },
  );

  let dados = null;
  try {
    dados = resposta.json();
  } catch (_) {
    // JSON inválido reprova a verificação de conteúdo.
  }
  const conteudoValido = Array.isArray(dados) && dados.length === 5
    && nomes.every((nome) => dados.some((servico) => servico
      && servico.name === nome && servico.group === 'seminario'
      && Array.isArray(servico.results) && servico.results.length === 50));

  check(resposta, {
    'HTTP 200': (resultado) => resultado.status === 200,
    'JSON com serviços e históricos esperados': () => conteudoValido,
  });
  sleep(1);
}
