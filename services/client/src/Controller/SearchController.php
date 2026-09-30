<?php

namespace App\Controller;

use App\Service\SearchResultFetcher;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\HttpKernel\Attribute\MapQueryParameter;
use Symfony\Component\Routing\Attribute\Route;

class SearchController extends AbstractController
{

    public function __construct(
        private SearchResultFetcher $searchResultFetcher,
    ) {
    }

    #[Route('/search', name: "search_results")]
    public function search(
        #[MapQueryParameter] string $query = '',
    ): Response {
        $results = $this->searchResultFetcher->FetchSearchResults($query);

        return $this->render('conference/show.html.twig', [
            'results' => $results,
        ]);
    }
}