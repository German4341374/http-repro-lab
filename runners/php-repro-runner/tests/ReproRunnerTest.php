<?php

declare(strict_types=1);

namespace HttpReproLab\Tests;

use HttpReproLab\ReproRunner;
use PHPUnit\Framework\TestCase;

final class ReproRunnerTest extends TestCase
{
    public function testRejectsEmptyUrl(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        (new ReproRunner())->execute('GET', '', [], null, 1000);
    }

    public function testRejectsUnsupportedScheme(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        (new ReproRunner())->execute('GET', 'file:///tmp/value', [], null, 1000);
    }

    public function testRejectsInvalidTimeout(): void
    {
        $this->expectException(\InvalidArgumentException::class);
        (new ReproRunner())->execute('GET', 'https://example.invalid', [], null, 0);
    }
}
