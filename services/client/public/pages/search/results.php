<?php if ($err !== null): ?>
    <h1>Hmm... looks like we've encountered an
        error on our part.</h1>
    <h2>Feel free to read through it if you're a nerd like us:</h2>
    <p><?= $err ?></p>
<?php else: ?>
    <h1>Found <?= sizeof($results->Results) ?>
        results for <i>
            <?= $results->InputtedQuery ?>
        </i>
        in <?= $results->GetDuration_s() ?> seconds:
    </h1>
    <?php foreach ($results->Results as $key => $value): ?>
        <li>
            <?= htmlspecialchars((string) $value, ENT_QUOTES, 'UTF-8') ?>
        </li>
    <?php endforeach; ?>
<?php endif ?>